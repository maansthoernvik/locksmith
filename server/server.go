package server

import (
	"context"
	"io"
	"net"

	"github.com/maansthoernvik/locksmith/env"
	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/protocol"
	"github.com/maansthoernvik/locksmith/server/connection"
	"github.com/maansthoernvik/locksmith/vault"
)

var logger *log.Logger

func init() {
	logLevel, _ := env.GetOptionalString(env.LOCKSMITH_LOG_LEVEL, env.LOCKSMITH_LOG_LEVEL_DEFAULT)
	logger = log.New(log.Translate(logLevel))
}

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
		logger.Error("Failed to start TCP acceptor")
		return err
	}
	logger.Info("Started locksmith on port:", locksmith.options.Port)

	locksmith.status = STARTED

	<-ctx.Done()
	logger.Info("Stopping locksmith")
	err = locksmith.tcpAcceptor.Stop()

	locksmith.status = STOPPED

	return err
}

// Incoming connections from the TCP acceptor come here first.
func (locksmith *Locksmith) handleConnection(conn net.Conn) {
	logger.Info("Connection accepted from:", conn.RemoteAddr().String())
	for {
		buffer := make([]byte, 257)
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				logger.Info("Connection", conn.RemoteAddr().String(),
					"closed by remote (EOF)")
			} else {
				logger.Info("Connection read error:", err)
			}

			// Connection has been closed, clean up client data
			locksmith.vault.Cleanup(conn.RemoteAddr().String())
			break
		}

		logger.Debug("Got message (", n, "chars)")
		logger.Debug("Buffer contains:", buffer)
		logger.Debug("Interesting part of the buffer:", buffer[:n])

		incomingMessage, err := protocol.DecodeServerMessage(buffer[:n])
		if err != nil {
			logger.Error("Decoding error, closing connection ("+
				conn.RemoteAddr().String()+"): ", err)
			break
		}

		locksmith.handleIncomingMessage(conn, incomingMessage)
	}
}

func (locksmith *Locksmith) handleIncomingMessage(
	conn net.Conn,
	incomingMessage *protocol.IncomingMessage,
) {
	switch incomingMessage.MessageType {
	case protocol.Acquire:
		locksmith.vault.Acquire(incomingMessage.LockTag, conn.RemoteAddr().String(), locksmith.acquireCallback(conn, incomingMessage.LockTag))
	case protocol.Release:
		locksmith.vault.Release(incomingMessage.LockTag, conn.RemoteAddr().String(), locksmith.releaseCallback(conn))
	default:
		logger.Error("Invalid message type")
	}
}

func (locksmith *Locksmith) acquireCallback(
	conn net.Conn,
	lockTag string,
) func(error) error {
	return func(err error) error {
		if err != nil {
			logger.Error("Got error in acquire callback:", err)
			conn.Close()
			return nil
		}

		logger.Debug("Notifying client of acquisition for lock tag", lockTag)
		_, writeErr := conn.Write(protocol.EncodeClientMessage(&protocol.OutgoingMessage{
			MessageType: protocol.Acquired,
			LockTag:     lockTag,
		}))
		if writeErr != nil {
			logger.Error("Failed to write to client:", writeErr)
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
			logger.Error("Got error in release callback:", err)
			conn.Close()
		}

		return nil
	}
}
