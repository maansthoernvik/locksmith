package connection

import (
	"fmt"
	"net"

	"github.com/maansthoernvik/locksmith/log"
)

type TCPAcceptor interface {
	Start() error
	Stop()
}

type TCPAcceptorImpl struct {
	port     uint16
	handler  func(net.Conn)
	listener net.Listener
	stop     chan interface{}
}

type TCPAcceptorOptions struct {
	Handler func(net.Conn)
	Port    uint16
}

func NewTCPAcceptor(options *TCPAcceptorOptions) TCPAcceptor {
	return &TCPAcceptorImpl{
		port:    options.Port,
		handler: options.Handler,
		stop:    make(chan interface{}),
	}
}

func (tcpAcceptor *TCPAcceptorImpl) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", tcpAcceptor.port))
	if err != nil {
		return err
	}

	tcpAcceptor.listener = listener
	go tcpAcceptor.startListener()

	return nil
}

func (tcpAcceptor *TCPAcceptorImpl) Stop() {
	log.GlobalLogger.Info("Stopping TCP acceptor")

	close(tcpAcceptor.stop)
	tcpAcceptor.listener.Close()
}

func (tcpAcceptor *TCPAcceptorImpl) startListener() {
	for {
		conn, err := tcpAcceptor.listener.Accept()
		if err != nil {
			select {
			case <-tcpAcceptor.stop:
				log.GlobalLogger.Info("Stopping accept loop gracefully")
				return
			default:
				log.GlobalLogger.Error(err)
			}
		} else {
			log.GlobalLogger.Debug("Listener accepted a connection")
			go tcpAcceptor.handler(conn)
		}
	}
}