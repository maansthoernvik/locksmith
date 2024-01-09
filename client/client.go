package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"

	"github.com/maansthoernvik/locksmith/env"
	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/protocol"
)

var logger *log.Logger

func init() {
	val, _ := env.GetOptionalString(env.LOCKSMITH_LOG_LEVEL, env.LOCKSMITH_LOG_LEVEL_DEFAULT)
	logger = log.New(log.Translate(val))
}

type Client interface {
	Acquire(lockTag string) error
	Release(lockTag string) error
	Connect() error
	Close()
}

type ClientOptions struct {
	Host       string
	Port       uint16
	TlsConfig  *tls.Config
	OnAcquired func(lockTag string)
}

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

func (clientImpl *clientImpl) Connect() (err error) {
	if clientImpl.tlsConfig != nil {
		logger.Info("Dialing (TLS)", clientImpl.host+":"+fmt.Sprint(clientImpl.port))
		clientImpl.conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", clientImpl.host, clientImpl.port), clientImpl.tlsConfig)
	} else {
		logger.Info("Dialing", clientImpl.host+":"+fmt.Sprint(clientImpl.port))
		clientImpl.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", clientImpl.host, clientImpl.port))
	}
	if err != nil {
		return err
	}
	logger.Info("Connected to", clientImpl.conn.RemoteAddr().String())

	go func(conn net.Conn) {
		defer conn.Close()
		for {
			buffer := make([]byte, 257)
			n, readErr := conn.Read(buffer)
			if readErr != nil {
				if readErr == io.EOF {
					logger.Info("Connection", conn.RemoteAddr().String(),
						"closed by remote (EOF)")
				} else {
					select {
					case <-clientImpl.stop:
						logger.Info("Stopping client connection gracefully")
					default:
						logger.Error("Connection read error:", readErr)
					}
				}

				break
			}

			clientMessage, decodeErr := protocol.DecodeClientMessage(buffer[:n])
			if decodeErr != nil {
				logger.Error("Failed to decode message:", decodeErr)
				continue
			}

			switch clientMessage.Type {
			case protocol.Acquired:
				clientImpl.onAcquired(clientMessage.LockTag)
			default:
				logger.Error("Client message type not recognized:", clientMessage.Type)
			}
		}
	}(clientImpl.conn)

	return nil
}

func (clientImpl *clientImpl) Close() {
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
