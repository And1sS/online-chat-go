package auth

import (
	"errors"
	"net/http"
	"strconv"
)

type Authorizer interface {
	Authorize(req *http.Request) (*Principal, error)
}

type NoOpAuthorizer struct{}

func (n *NoOpAuthorizer) Authorize(_ *http.Request) (*Principal, error) {
	return nil, errors.New("unimplemented yet")
}

type DummyAuthorizer struct{}

func (d *DummyAuthorizer) Authorize(_ *http.Request) (*Principal, error) {
	return &Principal{Id: strconv.Itoa(1)}, nil
}
