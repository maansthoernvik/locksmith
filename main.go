package main

import (
	"fmt"
	"io"
	"net"

	"github.com/maansthoernvik/locksmith/env"
	"github.com/maansthoernvik/locksmith/log"
	"go.uber.org/zap"
)

const ENV_HOST = "LS_HOST"
const ENV_PORT = "LS_PORT"

func main() {
	defer log.Logger.Sync()

	log.Logger.Info("Server listener starting")

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", env.RequiredString(ENV_HOST), env.RequiredUint16(ENV_PORT)))
	if err != nil {
		log.Logger.Fatal("Failed to start listener")
	}
	log.Logger.Info("Successfully started listener")

	log.Logger.Info("Awaiting client data...")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Logger.Info("Failed to accept an incoming connection", zap.Error(err))
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
			log.Logger.Info("Error reading from connection:", zap.String("error", err.Error()))
			if err == net.ErrClosed {
				log.Logger.Error("The error type indicated that the connection was closed, so terminating this goroutine")
				break
			} else if err == io.EOF {
				log.Logger.Error("EOF indicated, exiting")
				break
			}
			errorCount += 1
			continue
		}
		log.Logger.Info("Read message, sending reply \n", zap.Int("bytes", n), zap.ByteString("message", buffer))

		conn.Write([]byte{1, 1, 1, 1, 1})
	}

	log.Logger.Info("Exited connection read loop, terminating...")
	conn.Close()
}
