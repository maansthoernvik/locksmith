package connection

import (
	"fmt"
	"net"
)

type TCPAcceptor interface {
	Start(port uint16) error
}

type TCPAcceptorImpl struct {
	listener net.Listener
}

func NewTCPAcceptor() TCPAcceptor {
	return &TCPAcceptorImpl{}
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
			// TODO
		}
		go tcpAcceptor.handleConnection(conn)
	}
}

func (tcpAcceptor *TCPAcceptorImpl) handleConnection(conn net.Conn) {

}
