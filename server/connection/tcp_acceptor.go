package connection

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/maansthoernvik/locksmith/log"
)

type TCPAcceptor interface {
	Start() error
	Stop()
}

type TCPAcceptorOptions struct {
	Handler   func(net.Conn)
	Port      uint16
	TlsConfig *tls.Config
}

type tcpAcceptorImpl struct {
	port      uint16
	handler   func(net.Conn)
	tlsConfig *tls.Config
	listener  net.Listener
	stop      chan interface{}
}

func NewTCPAcceptor(options *TCPAcceptorOptions) TCPAcceptor {
	return &tcpAcceptorImpl{
		port:      options.Port,
		handler:   options.Handler,
		tlsConfig: options.TlsConfig,
		stop:      make(chan interface{}),
	}
}

func (tcpAcceptor *tcpAcceptorImpl) Start() (err error) {
	if tcpAcceptor.tlsConfig == nil {
		tcpAcceptor.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", tcpAcceptor.port))
		log.Info("Starting listener on port ", tcpAcceptor.port)
	} else {
		tcpAcceptor.listener, err = tls.Listen("tcp", fmt.Sprintf(":%d", tcpAcceptor.port), tcpAcceptor.tlsConfig)
		log.Info("Starting TLS listener on port ", tcpAcceptor.port)
	}
	if err != nil {
		return err
	}

	go tcpAcceptor.startListener()

	return nil
}

func (tcpAcceptor *tcpAcceptorImpl) Stop() {
	log.Info("Stopping TCP acceptor")
	close(tcpAcceptor.stop)
	tcpAcceptor.listener.Close()
}

func (tcpAcceptor *tcpAcceptorImpl) startListener() {
	defer tcpAcceptor.listener.Close()
	for {
		conn, err := tcpAcceptor.listener.Accept()
		if err != nil {
			select {
			case <-tcpAcceptor.stop:
				log.Info("Stopping accept loop gracefully")
			default:
				log.Error(err)
			}
			break
		}
		log.Debug("Listener accepted connection: ", conn.RemoteAddr().String())

		go func() {
			defer conn.Close()
			tcpAcceptor.handler(conn)
		}()
	}
}
