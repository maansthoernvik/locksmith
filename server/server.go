package server

import (
	"context"
	"io"
	"net"

	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/protocol"
	"github.com/maansthoernvik/locksmith/server/connection"
	"github.com/maansthoernvik/locksmith/vault"
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
	vault       vault.Vault
}

type LocksmithOptions struct {
	Port             uint16
	QueueType        vault.QueueType
	QueueConcurrency int
	QueueCapacity    int
}

func New(options *LocksmithOptions) *Locksmith {
	return &Locksmith{
		options: options,
		status:  STOPPED,
		vault: vault.NewVault(&vault.VaultOptions{
			QueueType:        options.QueueType,
			QueueConcurrency: options.QueueConcurrency,
			QueueCapacity:    options.QueueCapacity,
		}),
	}
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
		log.Error("Failed to start TCP acceptor")
		return err
	}
	log.Info("Started locksmith on port:", locksmith.options.Port)

	locksmith.status = STARTED

	<-ctx.Done()
	log.Info("Stopping locksmith")
	err = locksmith.tcpAcceptor.Stop()

	locksmith.status = STOPPED

	return err
}

// Incoming connections from the TCP acceptor come here first.
func (locksmith *Locksmith) handleConnection(conn net.Conn) {
	log.Info("Connection accepted from:", conn.RemoteAddr().String())
	for {
		buffer := make([]byte, 257)
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Info("Connection", conn.RemoteAddr().String(),
					"closed by remote (EOF)")
			} else {
				log.Error("Connection read error:", err)
			}

			// Connection has been closed, clean up client data
			locksmith.vault.Cleanup(conn.RemoteAddr().String())
			break
		}

		log.Debug("Got message (", n, "chars)")
		log.Debug("Buffer contains:", buffer)
		log.Debug("Interesting part of the buffer:", buffer[:n])

		incomingMessage, err := protocol.DecodeServerMessage(buffer[:n])
		if err != nil {
			log.Error("Decoding error, closing connection ("+
				conn.RemoteAddr().String()+"): ", err)
			break
		}

		locksmith.handleIncomingMessage(conn, incomingMessage)
	}
}

func (locksmith *Locksmith) handleIncomingMessage(
	conn net.Conn,
	serverMessage *protocol.ServerMessage,
) {
	switch serverMessage.Type {
	case protocol.Acquire:
		locksmith.vault.Acquire(serverMessage.LockTag, conn.RemoteAddr().String(), locksmith.acquireCallback(conn, serverMessage.LockTag))
	case protocol.Release:
		locksmith.vault.Release(serverMessage.LockTag, conn.RemoteAddr().String(), locksmith.releaseCallback(conn))
	default:
		log.Error("Invalid message type")
	}
}

func (locksmith *Locksmith) acquireCallback(
	conn net.Conn,
	lockTag string,
) func(error) error {
	return func(err error) error {
		if err != nil {
			log.Error("Got error in acquire callback:", err)
			conn.Close()
			return nil
		}

		log.Debug("Notifying client of acquisition for lock tag", lockTag)
		_, writeErr := conn.Write(protocol.EncodeClientMessage(&protocol.ClientMessage{
			Type:    protocol.Acquired,
			LockTag: lockTag,
		}))
		if writeErr != nil {
			log.Error("Failed to write to client:", writeErr)
			return writeErr
		}

		return nil
	}
}

func (locksmith *Locksmith) releaseCallback(
	conn net.Conn,
) func(error) error {
	return func(err error) error {
		if err != nil {
			log.Error("Got error in release callback:", err)
			conn.Close()
		}

		return nil
	}
}
