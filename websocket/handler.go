package websocket

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"online-chat-go/auth"
)

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewWsHandler(wss *WSConnections, authorizer auth.Authorizer) func(w http.ResponseWriter, r *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		principle, err := authorizer.Authorize(request)
		if err != nil {
			writer.WriteHeader(http.StatusUnauthorized)
			_, _ = writer.Write([]byte(err.Error()))
		}

		conn, err := websocketUpgrader.Upgrade(writer, request, nil)
		if err != nil {
			log.Println(err)
			return
		}

		connId := principle.Id
		wsconn := NewWsConnection(conn, messageHandler(connId), connectionCloseHandler(wss, connId))
		wss.AddConnection(connId, wsconn)
	}
}

func connectionCloseHandler(wss *WSConnections, id string) func(wsc *WSConnection, err error) {
	return func(wsc *WSConnection, err error) {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Printf("Unexpected error: %v", err)
		}
		_ = wss.RemoveConnection(id, wsc)
	}
}

func messageHandler(id string) func(messageType int, data []byte) {
	return func(messageType int, data []byte) {
		log.Printf("New message arrived from user: %s, msg: %s", id, data)
	}
}
