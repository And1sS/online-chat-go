package websocket

import (
	"github.com/docker/go-units"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type WsMessage struct {
	Type int
	Data []byte
}

type WsConfig struct {
	Timeout      time.Duration
	PingInterval time.Duration
	ReadLimit    int64
}

var defaultWsConfig = WsConfig{
	Timeout:      10 * time.Second,
	PingInterval: 1 * time.Second,
	ReadLimit:    64 * units.KB,
}

type wsConnection struct {
	writePump chan WsMessage
	readPump  chan WsMessage
	mut       *sync.Mutex // to prevent multiple goroutines from closing done channel
	done      chan bool
	config    WsConfig
	conn      *websocket.Conn
}

type WSConnection interface {
	WritePump() chan<- WsMessage

	ReadPump() <-chan WsMessage

	Done() <-chan bool

	Close() error
}

func (wsc *wsConnection) WritePump() chan<- WsMessage {
	return wsc.writePump
}

func (wsc *wsConnection) ReadPump() <-chan WsMessage {
	return wsc.readPump
}

func (wsc *wsConnection) Close() error {
	wsc.mut.Lock()
	defer wsc.mut.Unlock()

	select {
	case <-wsc.done:
		return nil

	default:
		close(wsc.done)
		return wsc.conn.Close()
	}
}

func (wsc *wsConnection) Done() <-chan bool {
	return wsc.done
}

func (wsc *wsConnection) runWriter() {
	ticker := time.NewTicker(wsc.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-wsc.done:
			return

		case <-ticker.C:
			if err := wsc.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}

		case msg := <-wsc.writePump:
			if err := wsc.write(msg.Type, msg.Data); err != nil {
				return
			}
		}
	}
}

func (wsc *wsConnection) write(msgType int, msgData []byte) error {
	_ = wsc.conn.SetWriteDeadline(time.Now().Add(wsc.config.Timeout))
	err := wsc.conn.WriteMessage(msgType, msgData)
	if err != nil {
		_ = wsc.Close()
	}

	return err
}

func (wsc *wsConnection) runReader() {
	for {
		select {
		case <-wsc.done:
			return

		default:
			if msgType, msgData, err := wsc.conn.ReadMessage(); err != nil {
				_ = wsc.Close()
				return
			} else {
				wsc.readPump <- WsMessage{Type: msgType, Data: msgData}
			}
		}
	}
}

func (wsc *wsConnection) setUp() {
	wsc.conn.SetReadLimit(wsc.config.ReadLimit)

	_ = wsc.conn.SetReadDeadline(time.Now().Add(wsc.config.Timeout))
	pongHandler := func(string) error {
		_ = wsc.conn.SetReadDeadline(time.Now().Add(wsc.config.Timeout))
		return nil
	}
	wsc.conn.SetPongHandler(pongHandler)

	go wsc.runWriter()
	go wsc.runReader()
}

func NewWsConnection(conn *websocket.Conn, config WsConfig) WSConnection {
	wsc := &wsConnection{
		conn:      conn,
		config:    config,
		writePump: make(chan WsMessage, 256),
		readPump:  make(chan WsMessage, 256),
		mut:       &sync.Mutex{},
		done:      make(chan bool),
	}

	wsc.setUp()
	return wsc
}
