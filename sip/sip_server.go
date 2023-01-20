package sip

import (
	"sync"
)

type RockSipSession struct {
}

type RockSipServer struct {
	SessionList sync.Map
}

func NewSipServer() *RockSipServer {
	sips := &RockSipServer{}

	return sips
}

func (sips *RockSipServer) Serve() (err error) {

	return err
}
