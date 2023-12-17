package server

import (
	"github.com/maansthoernvik/locksmith/env"
	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/server/connection"
)

type Locksmith struct {
	tcpAcceptor connection.TCPAcceptor
}

func New() *Locksmith {
	return &Locksmith{}
}

func (locksmith *Locksmith) Start() error {
	locksmith.tcpAcceptor = connection.NewTCPAcceptor()

	port, err := env.GetOptionalUint16(env.LOCKSMITH_PORT, env.LOCKSMITH_PORT_DEFAULT)
	if err != nil {
		log.GlobalLogger.Fatal("Failed to decode Locksmith port: ", err)
	}

	return locksmith.tcpAcceptor.Start(port)
}
