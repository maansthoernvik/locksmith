package main

import (
	"fmt"
	"time"

	"github.com/maansthoernvik/locksmith/connection"
	"github.com/maansthoernvik/locksmith/env"
)

const ENV_HOST = "LS_HOST"
const ENV_PORT = "LS_PORT"

func main() {
	fmt.Println("I'm a client")
	conn, err := connection.Connect(env.RequiredString(ENV_HOST), env.RequiredUint16(ENV_PORT))
	if err != nil {
		fmt.Println("Could not connect:", err)
		panic(err)
	}
	fmt.Println("Successfully connected")
	fmt.Println("Sending some bytes every few seconds")

	for {
		n, err := conn.Write([]byte{1, 2, 3})
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			break
		}
		fmt.Printf("Wrote %d bytes, awaiting reply \n", n)

		buffer := make([]byte, 5)
		n, err = conn.Read(buffer)
		if err != nil {
			fmt.Println("Got error while reading", err)
			break
		}
		fmt.Printf("Read %d bytes, message: %v \n", n, buffer)

		time.Sleep(5 * time.Second)
	}

	err = conn.Close()
	if err != nil {
		fmt.Println("Failed to close connection:", err)
	} else {
		fmt.Println("Successfully closed connection")
	}
}
