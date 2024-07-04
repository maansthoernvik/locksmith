// Package client provides a sample implementation of a Locksmith client.
package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"

	"github.com/maansthoernvik/locksmith/pkg/log"
	"github.com/maansthoernvik/locksmith/pkg/protocol"
)

// Client provides a simple interface for a Locksmith client.
type Client interface {
	Acquire(lockTag string) error
	Release(lockTag string) error
	Connect() error
	Close()
}

// ClientOptions to provide at client instantiation.
type ClientOptions struct {
	Host       string
	Port       uint16
	TlsConfig  *tls.Config
	OnAcquired func(lockTag string)
}

// Implements the Client interface.
type clientImpl struct {
	host       string
	port       uint16
	tlsConfig  *tls.Config
	onAcquired func(lockTag string)
	conn       net.Conn
	stop       chan interface{}
}

func NewClient(options *ClientOptions) Client {
	return &clientImpl{
		host:       options.Host,
		port:       options.Port,
		tlsConfig:  options.TlsConfig,
		onAcquired: options.OnAcquired,
		stop:       make(chan interface{}),
	}
}

// Connect to Locksmith, returning an error in case there is some connectivity error.
// Special note: if TLS version 13 is used, the Connect() function will not return an
// error, even if something is wrong, until the first client write is issues. This is
// because of how TLS 13 is implemented.
func (clientImpl *clientImpl) Connect() (err error) {
	if clientImpl.tlsConfig != nil {
		log.Info("Dialing (TLS) ", clientImpl.host+":"+fmt.Sprint(clientImpl.port))
		clientImpl.conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", clientImpl.host, clientImpl.port), clientImpl.tlsConfig)
	} else {
		log.Info("Dialing ", clientImpl.host+":"+fmt.Sprint(clientImpl.port))
		clientImpl.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", clientImpl.host, clientImpl.port))
	}
	if err != nil {
		return err
	}
	log.Info("Connected to ", clientImpl.conn.RemoteAddr().String())

	go func(conn net.Conn) {
		defer conn.Close()
		for {
			buffer := make([]byte, 257)
			n, readErr := conn.Read(buffer)
			if readErr != nil {
				if readErr == io.EOF {
					log.Info("Connection ", conn.RemoteAddr().String(),
						" closed by remote (EOF)")
				} else {
					select {
					case <-clientImpl.stop:
						log.Info("Stopping client connection gracefully")
					default:
						log.Error("Connection read error: ", readErr)
					}
				}

				break
			}

			clientMessage, decodeErr := protocol.DecodeClientMessage(buffer[:n])
			if decodeErr != nil {
				log.Error("Failed to decode message: ", decodeErr)
				continue
			}

			switch clientMessage.Type {
			case protocol.Acquired:
				clientImpl.onAcquired(clientMessage.LockTag)
			default:
				log.Error("Client message type not recognized: ", clientMessage.Type)
			}
		}
	}(clientImpl.conn)

	return nil
}

// Close disconnects from the Locksmith instance.
func (clientImpl *clientImpl) Close() {
	close(clientImpl.stop)
	clientImpl.conn.Close()
}

// Acquire the given lock tag.
// When the server responds, the onAcquired callback is called with the acquired lock tag.
func (clientImpl *clientImpl) Acquire(lockTag string) error {
	_, writeErr := clientImpl.conn.Write(
		protocol.EncodeServerMessage(
			&protocol.ServerMessage{Type: protocol.Acquire, LockTag: lockTag},
		),
	)

	return writeErr
}

// Release the given lock tag.
func (clientImpl *clientImpl) Release(lockTag string) error {
	_, writeErr := clientImpl.conn.Write(
		protocol.EncodeServerMessage(
			&protocol.ServerMessage{Type: protocol.Release, LockTag: lockTag},
		),
	)

	return writeErr
}
