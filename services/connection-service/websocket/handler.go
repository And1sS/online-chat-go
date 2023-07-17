package websocket

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"online-chat-go/auth"
	"online-chat-go/config"
)

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewWsHandler(wss *WSServer, authorizer auth.Authorizer, config *config.WsConfig, connHandler func(string, WSConnection)) func(w http.ResponseWriter, r *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		principal, err := authorizer.Authorize(request)
		if err != nil {
			writer.WriteHeader(http.StatusUnauthorized)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		conn, err := websocketUpgrader.Upgrade(writer, request, nil)
		if err != nil {
			log.Println(err)
			return
		}

		userId := principal.Id
		wsconn, err := NewWsConnection(conn, config)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("connected user via websocket with id: %s, connection id: %s\n", userId, wsconn.Id())

		go func() {
			_ = wss.AddConnection(userId, wsconn)
			defer func() { _ = wss.RemoveConnection(userId, wsconn) }()

			connHandler(userId, wsconn)
			log.Printf("disconnected user connected via websocket with id: %s, connection id: %s\n", userId, wsconn.Id())
		}()
	}
}
