package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"online-chat-go/util"
	"sync"
)

type WSError struct {
	msg string
}

func (err WSError) Error() string {
	return err.msg
}

type WSServer struct {
	connections map[string]*userWsConnections
	mut         *sync.RWMutex
}

func NewWSServer() *WSServer {
	conn := make(map[string]*userWsConnections)
	return &WSServer{
		connections: conn,
		mut:         &sync.RWMutex{},
	}
}

func (wss *WSServer) AddConnection(id string, conn WSConnection) {
	var userConns *userWsConnections
	var ok bool

	wss.mut.Lock()
	if userConns, ok = wss.connections[id]; !ok {
		userConns = newUserWsSessions(1)
		wss.connections[id] = userConns
	}
	wss.mut.Unlock()

	userConns.AddConnection(conn)
}

func (wss *WSServer) RemoveConnection(id string, conn WSConnection) error {
	wss.mut.Lock()
	defer wss.mut.Unlock()

	userConns, ok := wss.connections[id]

	if ok {
		remaining, err := userConns.RemoveConnection(conn)
		if remaining <= 0 {
			delete(wss.connections, id)
		}

		return err
	} else {
		return &WSError{fmt.Sprintf("No connections for id: %s", id)}
	}
}

func (wss *WSServer) SendTextMessage(id string, msg string) error {
	return wss.SendMessage(id, []byte(msg), websocket.TextMessage)
}

func (wss *WSServer) SendBinaryMessage(id string, msg []byte) error {
	return wss.SendMessage(id, msg, websocket.BinaryMessage)
}

func (wss *WSServer) SendMessage(id string, msgData []byte, msgType int) error {
	wss.mut.RLock()
	userConns, ok := wss.connections[id]
	wss.mut.RUnlock()

	if !ok {
		return &WSError{fmt.Sprintf("No connections for id: %s", id)}
	}

	userConns.ForAllConnections(
		func(conn WSConnection) {
			select {
			case <-conn.Done():
				return
			case conn.WritePump() <- WsMessage{Type: msgType, Data: msgData}:
				return
			}
		},
	)

	return nil
}

// internal holder struct for all opened ws connections of particular user
type userWsConnections struct {
	connections *[]WSConnection
	mut         *sync.RWMutex
}

func newUserWsSessions(initialCapacity int) *userWsConnections {
	s := make([]WSConnection, 0, initialCapacity)
	return &userWsConnections{
		connections: &s,
		mut:         &sync.RWMutex{},
	}
}

// AddConnection Does not perform contains check for speed, so same connection should not be added multiple times
func (u *userWsConnections) AddConnection(conn WSConnection) {
	u.mut.Lock()
	defer u.mut.Unlock()

	us := append(*u.connections, conn)
	u.connections = &us
}

func (u *userWsConnections) RemoveConnection(conn WSConnection) (int, error) {
	u.mut.Lock()
	defer u.mut.Unlock()

	err := util.RemoveSwapElem(u.connections, conn)
	return len(*u.connections), err
}

func (u *userWsConnections) ForAllConnections(block func(conn WSConnection)) {
	u.mut.RLock()
	defer u.mut.RUnlock()

	for _, connection := range *u.connections {
		go block(connection)
	}
}
