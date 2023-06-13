package websocket

import (
	"github.com/gorilla/websocket"
	"log"
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
	ReadLimit:    1024 * 64,
}

type wsConnection struct {
	writePump chan WsMessage
	readPump  chan WsMessage
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
	defer func() {
		ticker.Stop()
		_ = wsc.Close()
	}()

	for {
		select {
		case <-wsc.done:
			return

		case msg := <-wsc.writePump:
			err := wsc.writeWithDeadline(msg.Type, msg.Data)
			if err != nil {
				return
			}

		case <-ticker.C:
			err := wsc.writeWithDeadline(websocket.PingMessage, []byte{})
			if err != nil {
				return
			}
		}
	}
}

func (wsc *wsConnection) writeWithDeadline(msgType int, msgData []byte) error {
	_ = wsc.conn.SetWriteDeadline(time.Now().Add(wsc.config.Timeout))
	err := wsc.conn.WriteMessage(msgType, msgData)
	if err != nil {
		_ = wsc.Close()
	}

	return err
}

func (wsc *wsConnection) runReader() {
	wsc.conn.SetReadLimit(wsc.config.ReadLimit)
	_ = wsc.conn.SetReadDeadline(time.Now().Add(wsc.config.Timeout))

	pongHandler := func(string) error {
		_ = wsc.conn.SetReadDeadline(time.Now().Add(wsc.config.Timeout))
		return nil
	}
	wsc.conn.SetPongHandler(pongHandler)

	for {
		msgType, msgData, err := wsc.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			_ = wsc.Close()
			return
		}
		wsc.readPump <- WsMessage{Type: msgType, Data: msgData}
	}
}

func NewWsConnection(conn *websocket.Conn, config WsConfig) WSConnection {
	wsc := &wsConnection{
		conn:      conn,
		config:    config,
		writePump: make(chan WsMessage, 256),
		readPump:  make(chan WsMessage, 256),
		done:      make(chan bool),
	}
	go wsc.runWriter()
	go wsc.runReader()
	return wsc
}
