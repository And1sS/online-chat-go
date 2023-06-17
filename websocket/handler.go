package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"online-chat-go/auth"
	"time"
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

		timer := time.NewTicker(time.Second * 3)
		go func() {
			defer timer.Stop()

			for {
				select {
				case <-wsconn.Done():
					_ = wss.RemoveConnection(connId, wsconn)
					return

				case <-timer.C:
					wsconn.WritePump() <- WsMessage{Type: websocket.TextMessage, Data: []byte("Hello")}

				case msg := <-wsconn.ReadPump():
					log.Println(fmt.Sprintf("NEW MESSAGE FROM: %s, MSG: %s", connId, msg))
				}
			}
		}()
	}
}
