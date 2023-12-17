package connection

import (
	"fmt"
	"net"

	"github.com/maansthoernvik/locksmith/log"
)

type TCPAcceptor interface {
	Start(port uint16) error
	Stop()
}

type TCPAcceptorImpl struct {
	listener net.Listener
	handler  func(net.Conn)
	stop     chan interface{}
}

func NewTCPAcceptor(handler func(conn net.Conn)) TCPAcceptor {
	return &TCPAcceptorImpl{handler: handler, stop: make(chan interface{})}
}

func (tcpAcceptor *TCPAcceptorImpl) Start(port uint16) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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
