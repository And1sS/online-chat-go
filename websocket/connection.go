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

type wsConnection struct {
	writePump chan WsMessage
	readPump  chan WsMessage
	closePump chan bool
	conn      *websocket.Conn
}

type WSConnection interface {
	WritePump() chan<- WsMessage

	ReadPump() <-chan WsMessage

	ClosePump() <-chan bool

	Close() error
}

func (wsc *wsConnection) WritePump() chan<- WsMessage {
	return wsc.writePump
}

func (wsc *wsConnection) ReadPump() <-chan WsMessage {
	return wsc.readPump
}

func (wsc *wsConnection) Close() error {
	err := wsc.conn.Close()
	wsc.closePump <- true
	return err
}

func (wsc *wsConnection) ClosePump() <-chan bool {
	return wsc.closePump
}

func (wsc *wsConnection) runWriter() {
	ticker := time.NewTicker(time.Second * 1)
	defer func() {
		ticker.Stop()
		wsc.Close()
	}()

	for {
		select {
		case msg := <-wsc.writePump:
			log.Println("message written: ", msg)
			wsc.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			_ = wsc.conn.WriteMessage(msg.Type, msg.Data)
		case <-wsc.closePump:
			log.Println("writer closed")
			break
		case <-ticker.C:
			wsc.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if err := wsc.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				break
			}
		}
	}
}

func (wsc *wsConnection) runReader() {
	wsc.conn.SetReadLimit(1024)
	_ = wsc.conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	wsc.conn.SetPongHandler(func(string) error {
		_ = wsc.conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		log.Println("PONG RECEIVED")
		return nil
	})

	for {
		msgType, msgData, err := wsc.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			log.Println("connection closed with error: ", err)
			_ = wsc.Close()
			return
		}
		wsc.readPump <- WsMessage{Type: msgType, Data: msgData}
	}
}

func NewWsConnection(conn *websocket.Conn) WSConnection {
	wsc := &wsConnection{
		conn:      conn,
		writePump: make(chan WsMessage, 256),
		readPump:  make(chan WsMessage, 256),
		closePump: make(chan bool),
	}
	go wsc.runWriter()
	go wsc.runReader()
	return wsc
}
