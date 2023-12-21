package server

import (
	"context"
	"io"
	"net"

	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/server/connection"
)

type LocksmithStatus string

const (
	STARTED LocksmithStatus = "Started"
	STOPPED LocksmithStatus = "Stopped"
)

type Locksmith struct {
	tcpAcceptor connection.TCPAcceptor
	status      LocksmithStatus
	options     *LocksmithOptions
}

type LocksmithOptions struct {
	Port uint16
}

func New(options *LocksmithOptions) *Locksmith {
	return &Locksmith{options: options, status: STOPPED}
}

func (locksmith *Locksmith) Start(ctx context.Context) error {
	locksmith.tcpAcceptor = connection.NewTCPAcceptor(
		&connection.TCPAcceptorOptions{
			Port:    locksmith.options.Port,
			Handler: locksmith.handleConnection,
		},
	)
	err := locksmith.tcpAcceptor.Start()
	if err != nil {
		log.GlobalLogger.Error("Failed to start TCP acceptor")
		return err
	}
	log.GlobalLogger.Info("Started locksmith on port:", locksmith.options.Port)

	locksmith.status = STARTED

	<-ctx.Done()
	log.GlobalLogger.Info("Stopping locksmith")
	locksmith.tcpAcceptor.Stop()

	locksmith.status = STOPPED

	return nil
}

// Incoming connections from the TCP acceptor come here first.
func (locksmith *Locksmith) handleConnection(conn net.Conn) {
	log.GlobalLogger.Debug("Connection accepted from:", conn.RemoteAddr().String())
	for {
		buffer := make([]byte, 257)
		if _, err := conn.Read(buffer); err == io.EOF {
			log.GlobalLogger.Info("Connection", conn.RemoteAddr().String(), "closed by remote (EOF)")
			conn.Close()
			break
		}
		log.GlobalLogger.Info("Buffer contains:", buffer)
	}
}
