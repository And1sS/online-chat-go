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

func NewWsHandler(wss *WSServer, authorizer auth.Authorizer, config *config.WsConfig, connHandler func(WSConnection)) func(w http.ResponseWriter, r *http.Request) {
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

		connId := principal.Id
		wsconn := NewWsConnection(conn, config)
		if err != nil {
			log.Println(err)
			return
		}

		go func() {
			_ = wss.AddConnection(connId, wsconn)
			defer func() { _ = wss.RemoveConnection(connId, wsconn) }()

			connHandler(wsconn)
		}()
	}
}
