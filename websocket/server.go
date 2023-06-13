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

type WSConnections struct {
	connections map[string]*userWsConnections
	mut         *sync.RWMutex
}

func NewWSConnections() *WSConnections {
	conn := make(map[string]*userWsConnections)
	return &WSConnections{
		connections: conn,
		mut:         &sync.RWMutex{},
	}
}

func (wss *WSConnections) AddConnection(id string, conn WSConnection) {
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

func (wss *WSConnections) RemoveConnection(id string, conn WSConnection) error {
	var userConns *userWsConnections
	var ok bool

	wss.mut.RLock()
	userConns, ok = wss.connections[id]
	wss.mut.RUnlock()

	if ok {
		// TODO: Implement removal of holder from map
		return userConns.RemoveConnection(conn)
	} else {
		return &WSError{fmt.Sprintf("No connections for id: %s", id)}
	}
}

func (wss *WSConnections) SendTextMessage(id string, msg string) error {
	return wss.SendMessage(id, []byte(msg), websocket.TextMessage)
}

func (wss *WSConnections) SendBinaryMessage(id string, msg []byte) error {
	return wss.SendMessage(id, msg, websocket.BinaryMessage)
}

func (wss *WSConnections) SendMessage(id string, msgData []byte, msgType int) error {
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

			default:
				conn.WritePump() <- WsMessage{Type: msgType, Data: msgData}
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

func (u *userWsConnections) RemoveConnection(conn WSConnection) error {
	u.mut.Lock()
	defer u.mut.Unlock()

	return util.RemoveSwapElem(u.connections, conn)
}

func (u *userWsConnections) ForAllConnections(block func(conn WSConnection)) {
	u.mut.RLock()
	defer u.mut.RUnlock()

	for _, connection := range *u.connections {
		go block(connection)
	}
}
