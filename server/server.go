package server

import (
	"net"

	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/server/connection"
)

var logger = log.GlobalLogger

type Locksmith struct {
	tcpAcceptor connection.TCPAcceptor
	options     *LocksmithOptions
}

type LocksmithOptions struct {
	Port uint16
}

func New(options *LocksmithOptions) *Locksmith {
	return &Locksmith{options: options}
}

func (locksmith *Locksmith) Start() error {
	locksmith.tcpAcceptor = connection.NewTCPAcceptor(locksmith.handleConnection)
	return locksmith.tcpAcceptor.Start(locksmith.options.Port)
}

func (locksmith *Locksmith) handleConnection(conn net.Conn) {
	logger.Debug("Connection accepted", conn)
}
