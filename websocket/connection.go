package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"online-chat-go/util"
	"sync"
	"sync/atomic"
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

func (wss *WSConnections) AddConnection(id string, conn *WSConnection) {
	var userConns *userWsConnections
	var ok bool

	wss.mut.Lock()
	if conn.closed.Load() {
		wss.mut.Unlock()
		return
	}

	if userConns, ok = wss.connections[id]; !ok {
		userConns = newUserWsSessions(1)
		wss.connections[id] = userConns
	}
	wss.mut.Unlock()

	userConns.AddConnection(conn)
}

func (wss *WSConnections) RemoveConnection(id string, conn *WSConnection) error {
	var userConns *userWsConnections
	var ok bool

	wss.mut.Lock()
	userConns, ok = wss.connections[id]
	wss.mut.Unlock()

	// TODO: think about some sort of removal of empty user connections holders
	if ok {
		return userConns.RemoveConnection(conn)
	} else {
		return &WSError{fmt.Sprintf("No connections for id: %s", id)}
	}
}

func (wss *WSConnections) SendTextMessage(id string, msg string) []error {
	return wss.SendMessage(id, []byte(msg), websocket.TextMessage)
}

func (wss *WSConnections) SendBinaryMessage(id string, msg []byte) []error {
	return wss.SendMessage(id, msg, websocket.BinaryMessage)
}

func (wss *WSConnections) SendMessage(id string, msg []byte, messageType int) []error {
	wss.mut.RLock()
	userConns, ok := wss.connections[id]
	wss.mut.RUnlock()

	if !ok {
		return []error{&WSError{fmt.Sprintf("No connections for id: %s", id)}}
	}

	return userConns.ForAllConnections(
		func(wsConnection *WSConnection) error {
			return wsConnection.conn.WriteMessage(messageType, msg)
		},
	)
}

// internal holder struct for all opened ws connections of particular user
type userWsConnections struct {
	connections *[]*WSConnection
	mut         *sync.RWMutex
}

func newUserWsSessions(initialCapacity int) *userWsConnections {
	s := make([]*WSConnection, 0, initialCapacity)
	return &userWsConnections{
		connections: &s,
		mut:         &sync.RWMutex{},
	}
}

// AddConnection Does not perform contains check for speed, so same connection should not be added multiple times
func (u *userWsConnections) AddConnection(conn *WSConnection) {
	u.mut.Lock()
	defer u.mut.Unlock()

	us := append(*u.connections, conn)
	u.connections = &us
}

func (u *userWsConnections) RemoveConnection(conn *WSConnection) error {
	u.mut.Lock()
	defer u.mut.Unlock()

	return util.RemoveSwapElem(u.connections, conn)
}

func (u *userWsConnections) ForAllConnections(block func(conn *WSConnection) error) []error {
	u.mut.RLock()
	defer u.mut.RUnlock()

	var errors []error

	for _, connection := range *u.connections {
		err := block(connection)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

type WSConnection struct {
	conn               *websocket.Conn
	closed             atomic.Bool
	onMessageReceived  func(msgType int, data []byte)
	onConnectionClosed func(wsc *WSConnection, err error)
}

func NewWsConnection(
	conn *websocket.Conn,
	onMessageReceived func(msgType int, data []byte),
	onConnectionClosed func(wsc *WSConnection, err error),
) *WSConnection {
	wsc := &WSConnection{
		conn:               conn,
		closed:             atomic.Bool{},
		onMessageReceived:  onMessageReceived,
		onConnectionClosed: onConnectionClosed,
	}
	wsc.startMessageReader()
	return wsc
}

func (wsc *WSConnection) startMessageReader() {
	go func() {
		for {
			messageType, data, err := wsc.conn.ReadMessage()

			if err != nil {
				wsc.closed.Store(true)
				if wsc.onConnectionClosed != nil {
					wsc.onConnectionClosed(wsc, err)
				}
				break
			}

			if wsc.onMessageReceived != nil {
				wsc.onMessageReceived(messageType, data)
			}
		}
	}()
}
