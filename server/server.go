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
	Port uint16
}

func New(options *LocksmithOptions) *Locksmith {
	return &Locksmith{
		options: options,
		status:  STOPPED,
		vault:   vault.NewVault(&vault.VaultOptions{QueueType: vault.Single}),
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
		n, err := conn.Read(buffer)
		if err == io.EOF {
			log.GlobalLogger.Info("Connection", conn.RemoteAddr().String(),
				"closed by remote (EOF)")
			conn.Close()
			break
		} else if err != nil {
			log.GlobalLogger.Info("Connection closed:", err)
			conn.Close()
			break
		}

		log.GlobalLogger.Debug("Got message (", n, "chars)")
		log.GlobalLogger.Debug("Buffer contains:", buffer)
		log.GlobalLogger.Debug("Interesting part of the buffer:", buffer[:n])

		incomingMessage, err := protocol.DecodeServerMessage(buffer[:n])
		if err != nil {
			log.GlobalLogger.Error("Decoding error, closing connection ("+
				conn.RemoteAddr().String()+"): ", err)
			conn.Close()
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
		log.GlobalLogger.Error("Invalid message type")
	}
}

func (locksmith *Locksmith) acquireCallback(
	conn net.Conn,
	lockTag string,
) func(error) {
	return func(err error) {
		if err != nil {
			log.GlobalLogger.Error("Got error in acquire callback:", err)
			conn.Close()
			return
		}

		log.GlobalLogger.Debug("Notifying client of acquisition for lock tag", lockTag)
		_, writeErr := conn.Write(protocol.EncodeClientMessage(&protocol.OutgoingMessage{
			MessageType: protocol.Acquired,
			LockTag:     lockTag,
		}))
		if writeErr != nil {
			log.GlobalLogger.Error("Failed to write to client:", writeErr)
		}
	}
}

func (locksmith *Locksmith) releaseCallback(
	conn net.Conn,
) func(error) {
	return func(err error) {
		if err != nil {
			log.GlobalLogger.Error("Got error in release callback:", err)
			conn.Close()
		}
	}
}
