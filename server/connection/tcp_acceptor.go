package connection

import (
	"fmt"
	"net"

	"github.com/maansthoernvik/locksmith/log"
)

var logger = log.GlobalLogger

type TCPAcceptor interface {
	Start(port uint16) error
}

type TCPAcceptorImpl struct {
	listener net.Listener
	handler  func(net.Conn)
}

func NewTCPAcceptor(handler func(conn net.Conn)) TCPAcceptor {
	return &TCPAcceptorImpl{handler: handler}
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

func (tcpAcceptor *TCPAcceptorImpl) startListener() {
	for {
		conn, err := tcpAcceptor.listener.Accept()
		if err != nil {
			logger.Error(err)
		} else {
			go tcpAcceptor.handler(conn)
		}
	}
}
