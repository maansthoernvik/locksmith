package server

import (
	"net"

	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/server/connection"
)

var logger = log.GlobalLogger

type Locksmith struct {
	tcpAcceptor connection.TCPAcceptor
}

func New() *Locksmith {
	return &Locksmith{}
}

func (locksmith *Locksmith) Start(port uint16) error {
	locksmith.tcpAcceptor = connection.NewTCPAcceptor(locksmith.handleConnection)
	return locksmith.tcpAcceptor.Start(port)
}

func (locksmith *Locksmith) handleConnection(conn net.Conn) {
	logger.Debug("Connection accepted", conn)
}
