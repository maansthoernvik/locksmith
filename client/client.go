package client

import (
	"fmt"
	"io"
	"net"

	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/protocol"
)

type Client interface {
	Acquire(lockTag string) error
	Release(lockTag string) error
	Start() error
	Stop()
}

type ClientOptions struct {
	Host       string
	Port       uint16
	OnAcquired func(lockTag string)
}

type clientImpl struct {
	host       string
	port       uint16
	onAcquired func(lockTag string)
	conn       net.Conn
	stop       chan interface{}
}

func NewClient(options *ClientOptions) Client {
	return &clientImpl{
		host:       options.Host,
		port:       options.Port,
		onAcquired: options.OnAcquired,
		stop:       make(chan interface{}),
	}
}

func (clientImpl *clientImpl) Start() error {
	conn, dialErr := net.Dial("tcp", fmt.Sprintf("%s:%d", clientImpl.host, clientImpl.port))
	if dialErr != nil {
		return dialErr
	}
	clientImpl.conn = conn

	go func() {
		for {
			buffer := make([]byte, 257)
			n, readErr := conn.Read(buffer)
			if readErr != nil {
				if readErr == io.EOF {
					log.Info("Connection", conn.RemoteAddr().String(),
						"closed by remote (EOF)")
				} else {
					select {
					case <-clientImpl.stop:
						log.Info("Stopping client connection gracefully")
						return
					default:
						log.Error("Connection read error:", readErr)
					}
				}

				return
			}

			clientMessage, decodeErr := protocol.DecodeClientMessage(buffer[:n])
			if decodeErr != nil {
				log.Error("Failed to decode message:", decodeErr)
				continue
			}

			switch clientMessage.Type {
			case protocol.Acquired:
				clientImpl.onAcquired(clientMessage.LockTag)
			default:
				log.Error("Client message type not recognized:", clientMessage.Type)
			}
		}
	}()

	return nil
}

func (clientImpl *clientImpl) Stop() {
	close(clientImpl.stop)
	clientImpl.conn.Close()
}

func (clientImpl *clientImpl) Acquire(lockTag string) error {
	_, writeErr := clientImpl.conn.Write(
		protocol.EncodeServerMessage(
			&protocol.ServerMessage{Type: protocol.Acquire, LockTag: lockTag},
		),
	)

	return writeErr
}

func (clientImpl *clientImpl) Release(lockTag string) error {
	_, writeErr := clientImpl.conn.Write(
		protocol.EncodeServerMessage(
			&protocol.ServerMessage{Type: protocol.Release, LockTag: lockTag},
		),
	)

	return writeErr
}
