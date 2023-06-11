package auth

import (
	"errors"
	"math/rand"
	"net/http"
	"strconv"
)

type Authorizer interface {
	Authorize(req *http.Request) (*Principle, error)
}

type NoOpAuthenticator struct{}

func (n *NoOpAuthenticator) Authorize(_ *http.Request) (*Principle, error) {
	return nil, errors.New("unimplemented yet")
}

type DummyAuthenticator struct{}

func (d *DummyAuthenticator) Authorize(_ *http.Request) (*Principle, error) {
	return &Principle{Id: strconv.Itoa(rand.Int())}, nil
}
