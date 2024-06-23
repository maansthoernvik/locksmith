// This package implements a simple CLI for Locksmith.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/maansthoernvik/locksmith/client"
)

const USAGE = `Starts a session towards a Locksmith instance using the sample Locksmith
client implementation.`
const COMMANDS = `Session started, the following commands are supported:

acquire [lock]
release [lock]`

var host string
var port uint

func main() {
	flag.StringVar(&host, "host", "localhost", "Locksmith hostname or IP address.")
	flag.UintVar(&port, "port", 9000, "Locksmith port number.")

	flag.Usage = func() {
		fmt.Println(USAGE)
		fmt.Println()
		flag.CommandLine.PrintDefaults()
	}

	flag.Parse()

	fmt.Println("Starting Locksmith shell...")
	acquiredChan := make(chan interface{}, 1)
	c, err := getClient(acquiredChan)
	if err != nil {
		fmt.Println("Error creating client:", err)
		return
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Exiting")
		c.Close()
		fmt.Println("CLOSED:", fmt.Sprintf("%s:%d", host, port))
		os.Exit(0)
	}()

	fmt.Println("CONNECTED:", fmt.Sprintf("%s:%d", host, port))
	fmt.Println("")
	fmt.Println(COMMANDS)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Encountered error:", err)
			sigChan <- syscall.SIGINT
			return
		}

		cleanedText := strings.Split(text[:len(text)-1], " ")
		switch cleanedText[0] {
		case "exit":
			sigChan <- syscall.SIGINT
		case "acquire":
			if len(cleanedText) != 2 {
				fmt.Println("> expected 'acquire' followed by a lock")
				sigChan <- syscall.SIGINT
				return
			}
			lock := cleanedText[1]
			err := c.Acquire(lock)
			if err != nil {
				fmt.Println("Encountered error:", err)
				sigChan <- syscall.SIGINT
				return
			}
			select {
			case <-time.After(2 * time.Second):
				fmt.Println("Timed out waiting for acquired signal")
				sigChan <- syscall.SIGINT
				return
			case <-acquiredChan:
				//noop
			}
		case "release":
			if len(cleanedText) != 2 {
				fmt.Println("> expected 'release' followed by a lock")
				sigChan <- syscall.SIGINT
				return
			}
			lock := cleanedText[1]
			err := c.Release(lock)
			if err != nil {
				fmt.Println("Encountered error:", err)
				sigChan <- syscall.SIGINT
				return
			}
		default:
			if cleanedText[0] != "" {
				fmt.Println("Did not recognize command:", cleanedText[0])
			}
		}
	}
}

func getClient(acquiredChan chan interface{}) (client.Client, error) {
	c := client.NewClient(&client.ClientOptions{
		Host: host,
		Port: uint16(port),
		OnAcquired: func(lock string) {
			fmt.Println("acquired ", lock)
			acquiredChan <- nil
		},
	})
	if err := c.Connect(); err != nil {
		return nil, err
	}

	return c, nil
}
