package connection

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/maansthoernvik/locksmith/client"
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
	defer tcpAcceptor.Stop()

	go func(wg *sync.WaitGroup) {
		<-time.After(1 * time.Second)
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

func TestTcpAcceptor_ClientEvictedNoTls(t *testing.T) {
	cert, err := tls.LoadX509KeyPair("testcerts/testCert.pem", "testcerts/testKey.pem")
	if err != nil {
		t.Error("Error when loading cert and key pair", err)
	}

	clientCaCert, err := os.ReadFile("testcerts/rootCACert.pem")
	if err != nil {
		t.Error("Failed to read client CA cert:", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(clientCaCert)
	tcpAcceptor := NewTCPAcceptor(&TCPAcceptorOptions{
		Handler: func(conn net.Conn) {
			defer conn.Close()
			for {
				_, err := conn.Read(make([]byte, 25))
				t.Log("Got bytes from client...")
				if err != nil {
					t.Log("Got expected error reading:", err)
					break
				}
				t.Error("Did not get an expected error while reading...")
				break //nolint
			}
			wg.Done()
		},
		Port: 30000,
		TlsConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    pool,
			MinVersion:   tls.VersionTLS13,
		},
	})

	err = tcpAcceptor.Start()
	if err != nil {
		t.Error("Error when starting tcp acceptor:", err)
	}
	defer tcpAcceptor.Stop()

	c := client.NewClient(&client.ClientOptions{
		Host: "localhost",
		Port: 30000,
	})

	err = c.Connect()
	if err != nil {
		t.Error("Error when connecting client:", err)
	}
	defer c.Close()

	c.Acquire("abc") //nolint
	t.Log("Awaiting listener read...")
	wg.Wait()
}

func TestTcpAcceptor_MutualTls(t *testing.T) {
	cert, err := tls.LoadX509KeyPair("testcerts/testCert.pem", "testcerts/testKey.pem")
	if err != nil {
		t.Error("Error when loading cert and key pair", err)
	}

	clientCaCert, err := os.ReadFile("testcerts/rootCACert.pem")
	if err != nil {
		t.Error("Failed to read client CA cert:", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(clientCaCert)
	tcpAcceptor := NewTCPAcceptor(&TCPAcceptorOptions{
		Handler: func(conn net.Conn) {
			defer conn.Close()
			for {
				_, err := conn.Read(make([]byte, 25))
				t.Log("Got bytes from client...")
				if err != nil {
					t.Error("Got expected error reading:", err)
					break
				}
				t.Log("No error while reading, quitting connection loop")
				break //nolint
			}
			wg.Done()
		},
		Port: 30000,
		TlsConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    pool,
			MinVersion:   tls.VersionTLS13,
		},
	})

	err = tcpAcceptor.Start()
	if err != nil {
		t.Error("Error when starting tcp acceptor:", err)
	}
	defer tcpAcceptor.Stop()

	clientCert, err := tls.LoadX509KeyPair("testcerts/testCert.pem", "testcerts/testKey.pem")
	if err != nil {
		t.Error(err)
	}
	c := client.NewClient(&client.ClientOptions{
		Host: "localhost",
		Port: 30000,
		TlsConfig: &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      pool,
			MinVersion:   tls.VersionTLS13,
		},
	})

	err = c.Connect()
	if err != nil {
		t.Error("Error when connecting client:", err)
	}
	defer c.Close()

	c.Acquire("abc") //nolint
	t.Log("Awaiting listener read...")
	wg.Wait()
}
