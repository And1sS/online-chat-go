package websocket

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"online-chat-go/util"
	"runtime"
)

type WSServer struct {
	connections        *util.SafeMap[string, *userWsConnections]
	onUserConnected    func(id string)
	onUserDisconnected func(id string)
}

func NewWSServer() *WSServer {
	return &WSServer{connections: util.NewSafeMap[string, *userWsConnections]()}
}

func (wss *WSServer) SetOnUserConnected(callback func(id string)) {
	wss.onUserConnected = callback
}

func (wss *WSServer) SetOnUserDisconnected(callback func(id string)) {
	wss.onUserDisconnected = callback
}

func (wss *WSServer) AddConnection(id string, conn WSConnection) error {
	for {
		userConns, created := wss.connections.ComputeIfAbsent(id, newSingleUserWsConnection)
		err := userConns.AddConnection(conn)

		if err == nil {
			if created && wss.onUserConnected != nil {
				wss.onUserConnected(id)
			}

			return nil
		} else if _, ok := err.(*DestroyedUConnUsageError); ok {
			runtime.Gosched() // We can't add connection to destroyed holder, so we need to retry later
		} else {
			return err
		}
	}
}

func (wss *WSServer) RemoveConnection(id string, conn WSConnection) error {
	userConns, ok := wss.connections.Get(id)
	if !ok {
		return errors.New(fmt.Sprintf("No connections for id: %s", id))
	}

	destroyed, err := userConns.RemoveConnection(conn)
	if destroyed {
		wss.connections.Delete(id)
		if wss.onUserDisconnected != nil {
			wss.onUserDisconnected(id)
		}
	}

	return err
}

func (wss *WSServer) SendTextMessage(id string, msg string) error {
	return wss.SendMessage(id, []byte(msg), websocket.TextMessage)
}

func (wss *WSServer) SendBinaryMessage(id string, msg []byte) error {
	return wss.SendMessage(id, msg, websocket.BinaryMessage)
}

func (wss *WSServer) SendMessage(id string, msgData []byte, msgType int) error {
	userConns, ok := wss.connections.Get(id)
	if !ok {
		return errors.New(fmt.Sprintf("No connections for id: %s", id))
	}

	return userConns.ForAllConnections(
		func(conn WSConnection) {
			select {
			case <-conn.Done():
				return
			case conn.WritePump() <- WsMessage{Type: msgType, Data: msgData}:
				return
			}
		},
	)
}
