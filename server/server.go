// Package server ties together the Locksmith server logic.
package server

import (
	"context"
	"crypto/tls"
	"io"
	"net"

	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/protocol"
	"github.com/maansthoernvik/locksmith/server/connection"
	"github.com/maansthoernvik/locksmith/vault"
)

// Locksmith is the root level object containing the implementation of the Locksmith server.
type Locksmith struct {
	tcpAcceptor connection.TCPAcceptor
	vault       vault.Vault
}

// LocksmithOptions exposes the possible options to pass to a new Locksmith instance.
type LocksmithOptions struct {
	// Denotes the port which will listen for incoming connections.
	Port uint16
	// Selects the type of queue layer the vault will use.
	QueueType vault.QueueType
	// Sets the number of synchronization threads, the higher the number the less the chance of congestion.
	QueueConcurrency int
	// Determines the buffer size of each synchronization thread, after the buffer limit is reached, calls
	// to the queue layer will block until the congestion is resolved.
	QueueCapacity int
	// TLS configuration for the TCP acceptor.
	TlsConfig *tls.Config
}

func New(options *LocksmithOptions) *Locksmith {
	locksmith := &Locksmith{
		vault: vault.NewVault(&vault.VaultOptions{
			QueueType:        options.QueueType,
			QueueConcurrency: options.QueueConcurrency,
			QueueCapacity:    options.QueueCapacity,
		}),
	}
	locksmith.tcpAcceptor = connection.NewTCPAcceptor(&connection.TCPAcceptorOptions{
		Handler:   locksmith.handleConnection,
		Port:      options.Port,
		TlsConfig: options.TlsConfig,
	})

	return locksmith
}

// Blocking call! Starts the Locksmith instance. Call Stop() to stop the instance.
func (locksmith *Locksmith) Start(ctx context.Context) error {
	err := locksmith.tcpAcceptor.Start()
	if err != nil {
		log.Error("Failed to start TCP acceptor")
		return err
	}
	log.Info("Started locksmith")

	<-ctx.Done()
	log.Info("Stopping locksmith")
	locksmith.tcpAcceptor.Stop()

	return err
}

// Handler for connections accepted by the TCP acceptor. This function contains
// a connection loop which only ends upon the client connection encountering an
// error, either due to a problem or shutdown of the client connection. Gotten
// messages will be attempted to be decoded, if decoding fails the loop is
// broken and the client connection disconnected.
func (locksmith *Locksmith) handleConnection(conn net.Conn) {
	log.Info("Connection accepted from: ", conn.RemoteAddr().String())
	for {
		buffer := make([]byte, 257)
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Info("Connection ", conn.RemoteAddr().String(),
					" closed by remote (EOF)")
			} else {
				log.Error("Connection read error: ", err)
			}

			// Connection error, clean up client data
			locksmith.vault.Cleanup(conn.RemoteAddr().String())
			break
		}

		log.Debug("Got message (", n, " chars)")
		log.Debug("Buffer contains: ", buffer)
		log.Debug("Interesting part of the buffer: ", buffer[:n])

		incomingMessage, err := protocol.DecodeServerMessage(buffer[:n])
		if err != nil {
			log.Error("Decoding error, closing connection ("+
				conn.RemoteAddr().String()+"): ", err)
			break
		}

		locksmith.handleIncomingMessage(conn, incomingMessage)
	}
}

// After decoding, this function determines the handling of the decoded
// message.
func (locksmith *Locksmith) handleIncomingMessage(
	conn net.Conn,
	serverMessage *protocol.ServerMessage,
) {
	switch serverMessage.Type {
	case protocol.Acquire:
		locksmith.vault.Acquire(
			serverMessage.LockTag,
			conn.RemoteAddr().String(),
			locksmith.acquireCallback(conn, serverMessage.LockTag),
		)
	case protocol.Release:
		locksmith.vault.Release(
			serverMessage.LockTag,
			conn.RemoteAddr().String(),
			locksmith.releaseCallback(conn),
		)
	default:
		log.Error("Invalid message type")
	}
}

// Returns a callback function to call once a lock has been acquired, to send
// feedback down the client connection. If the callback is called with an error,
// the client has misbehaved in some way and needs to be disconnected.
func (locksmith *Locksmith) acquireCallback(
	conn net.Conn,
	lockTag string,
) func(error) error {
	return func(err error) error {
		if err != nil {
			log.Error("Got error in acquire callback: ", err)
			conn.Close()
			return nil
		}

		log.Debug("Notifying client of acquisition for lock tag ", lockTag)
		_, writeErr := conn.Write(protocol.EncodeClientMessage(&protocol.ClientMessage{
			Type:    protocol.Acquired,
			LockTag: lockTag,
		}))
		if writeErr != nil {
			log.Error("Failed to write to client: ", writeErr)
			return writeErr
		}

		return nil
	}
}

// Returns a callback function to call once a lock has been released. If the
// callback is called with an error, the client has misbehaved in some way and
// needs to be disconnected.
func (locksmith *Locksmith) releaseCallback(
	conn net.Conn,
) func(error) error {
	return func(err error) error {
		if err != nil {
			log.Error("Got error in release callback: ", err)
			conn.Close()
		}

		return nil
	}
}
