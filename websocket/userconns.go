package websocket

import (
	"online-chat-go/util"
	"sync"
)

type DestroyedUConnUsageError struct {
	msg string
}

func (err *DestroyedUConnUsageError) Error() string {
	return err.msg
}

// internal holder struct for all opened ws connections of particular user
type userWsConnections struct {
	connections *[]WSConnection
	mut         *sync.RWMutex
	destroyed   bool
}

func newSingleUserWsConnection() *userWsConnections {
	return newUserWsConnections(1)
}

func newUserWsConnections(initialCapacity int) *userWsConnections {
	s := make([]WSConnection, 0, initialCapacity)
	return &userWsConnections{
		connections: &s,
		mut:         &sync.RWMutex{},
		destroyed:   false,
	}
}

// AddConnection Does not perform contains check for speed, so same connection should not be added multiple times
func (u *userWsConnections) AddConnection(conn WSConnection) error {
	u.mut.Lock()
	defer u.mut.Unlock()

	if u.destroyed == true {
		return &DestroyedUConnUsageError{"Tying to add connection to destroyed holder"}
	}

	us := append(*u.connections, conn)
	u.connections = &us
	return nil
}

func (u *userWsConnections) RemoveConnection(conn WSConnection) (bool, error) {
	u.mut.Lock()
	defer u.mut.Unlock()

	err := util.RemoveSwapElem(u.connections, conn)
	if len(*u.connections) == 0 {
		u.destroyed = true
	}

	return u.destroyed, err
}

func (u *userWsConnections) ForAllConnections(block func(conn WSConnection)) error {
	u.mut.RLock()
	defer u.mut.RUnlock()

	if u.destroyed {
		return &DestroyedUConnUsageError{"Trying to run function over destroyed connections holder"}
	}

	for _, connection := range *u.connections {
		go block(connection)
	}
	return nil
}
