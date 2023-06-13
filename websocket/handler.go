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
		wsconn := NewWsConnection(conn, defaultWsConfig)
		wss.AddConnection(connId, wsconn)

		go func() {
			for {
				select {
				case <-wsconn.Done():
					_ = wss.RemoveConnection(connId, wsconn)
					return
				case msg := <-wsconn.ReadPump():
					log.Println("NEW MESSAGE FROM: %s, MSG: %s", connId, msg)
				}
			}
		}()
	}
}
