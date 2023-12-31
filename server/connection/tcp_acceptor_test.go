package connection

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestTcpAcceptor_AcceptConnections(t *testing.T) {
	// Start a TCP acceptor with a handler function, expect the function to be
	// called on establishing a connection.
	wg := &sync.WaitGroup{}
	wg.Add(1)

	handled_connection := false
	tcpAcceptor := NewTCPAcceptor(&TCPAcceptorOptions{
		Handler: func(conn net.Conn) {
			t.Log("Connection handler called!")
			handled_connection = true
			defer conn.Close()
			wg.Done()
		},
		Port: 30000,
	})
	// Use a port that isn't likely in use nor in a range that could be
	// privileged.
	err := tcpAcceptor.Start()
	if err != nil {
		t.Error("Failed to start TCP acceptor %w", err)
	}

	go func(wg *sync.WaitGroup) {
		<-time.After(5 * time.Second)
		if !handled_connection {
			t.Error("Did not handle connection...")
			wg.Done()
		}
	}(wg)

	client_conn, err := net.Dial("tcp", "localhost:30000")
	if err != nil {
		t.Error("Error when dialing localhost:30000")
	}
	defer client_conn.Close()

	wg.Wait()
}
