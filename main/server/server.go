package main

import (
	"fmt"
	"io"
	"net"

	"github.com/maansthoernvik/locksmith/lib/env"
)

const ENV_HOST = "LS_HOST"
const ENV_PORT = "LS_PORT"

func main() {
	fmt.Println("I'm a server")

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", env.RequiredString(ENV_HOST), env.RequiredUint16(ENV_PORT)))
	if err != nil {
		fmt.Println("Failed to start listener")
		panic(err)
	}
	fmt.Println("Successfully started listener")

	fmt.Println("Awaiting client data...")
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept an incoming connection", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	errorCount := 0
	for errorCount < 3 {
		buffer := make([]byte, 5)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from connection:", err.Error())
			if err == net.ErrClosed {
				fmt.Println("The error type indicated that the connection was closed, so terminating this goroutine")
				break
			} else if err == io.EOF {
				fmt.Println("EOF indicated, exiting")
				break
			}
			errorCount += 1
			continue
		}
		fmt.Printf("Read %d bytes, message %v, sending reply \n", n, buffer)

		conn.Write([]byte{1, 1, 1, 1, 1})
	}

	fmt.Println("Exited connection read loop, terminating...")
	conn.Close()
}
