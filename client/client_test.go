package client

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/maansthoernvik/locksmith/protocol"
)

func Test_ClientLifecycle(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:30000")
	if err != nil {
		t.Fatal("Failed to start listener:", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			t.Log("Client established connection")
			wg.Done()

			for {
				t.Log("Reading from client connection")
				buffer := make([]byte, 100)
				_, err = conn.Read(buffer)
				t.Log("Read from client connection")
				if err == io.EOF {
					t.Log("Client closed connection")
					wg.Done()
					return
				}
			}
		}
	}()

	client := NewClient(&ClientOptions{Host: "localhost", Port: 30000})
	startErr := client.Connect()
	if err != nil {
		t.Fatal("Failed to start client:", startErr)
	}
	client.Close()

	wg.Wait()

	listener.Close()
}

func Test_ClientAcquireRelease(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:30000")
	if err != nil {
		t.Fatal("Failed to start listener:", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			t.Log("Client established connection")

			for {
				t.Log("Reading from client connection")
				buffer := make([]byte, 100)
				n, err := conn.Read(buffer)
				t.Log("Read from client connection")
				if err == io.EOF {
					t.Log("Client closed connection")
					return
				}

				serverMessage, err := protocol.DecodeServerMessage(buffer[:n])
				if err != nil {
					t.Error("Error decoding server message:", err)
					return
				}

				if serverMessage.Type == protocol.Acquire {
					t.Log("Acquire received")
					wg.Done()
				} else if serverMessage.Type == protocol.Release {
					t.Log("Release received")
					wg.Done()
				}
			}
		}
	}()

	client := NewClient(&ClientOptions{Host: "localhost", Port: 30000})
	startErr := client.Connect()
	if err != nil {
		t.Fatal("Failed to start client:", startErr)
	}
	_ = client.Acquire("123")
	time.Sleep(1 * time.Millisecond)
	_ = client.Release("123")

	wg.Wait()

	client.Close()
	listener.Close()
}

func Test_ClientOnAcquired(t *testing.T) {
	EXPECTED_LOCK_TAG := "locktag"

	listener, err := net.Listen("tcp", "localhost:30000")
	if err != nil {
		t.Fatal("Failed to start listener:", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			t.Log("Client established connection")

			for {
				t.Log("Reading from client connection")
				buffer := make([]byte, 100)
				n, err := conn.Read(buffer)
				t.Log("Read from client connection")
				if err == io.EOF {
					t.Log("Client closed connection")
					return
				}

				serverMessage, err := protocol.DecodeServerMessage(buffer[:n])
				if err != nil {
					t.Error("Error decoding server message:", err)
					return
				}

				if serverMessage.Type == protocol.Acquire {
					t.Log("Acquire received")
					wg.Done()

					_, err := conn.Write(protocol.EncodeClientMessage(
						&protocol.ClientMessage{Type: protocol.Acquired, LockTag: serverMessage.LockTag},
					))
					if err != nil {
						t.Error("Got error on write:", err)
					}
				}
			}
		}
	}()

	client := NewClient(&ClientOptions{Host: "localhost", Port: 30000, OnAcquired: func(lockTag string) {
		if lockTag == EXPECTED_LOCK_TAG {
			t.Log("OnAcquired called")
			wg.Done()
		}
	}})
	startErr := client.Connect()
	if err != nil {
		t.Fatal("Failed to start client:", startErr)
	}
	_ = client.Acquire(EXPECTED_LOCK_TAG)

	wg.Wait()
	client.Close()
	listener.Close()
}
