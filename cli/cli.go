package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/maansthoernvik/locksmith/client"
)

var host string
var port uint

func main() {
	flag.StringVar(&host, "host", "localhost", "Locksmith hostname or IP address.")
	flag.UintVar(&port, "port", 9000, "Locksmith port number.")

	flag.Parse()

	fmt.Println("Starting Locksmith shell...")
	c, err := getClient()
	if err != nil {
		fmt.Println("Error creating client:", err)
		return
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		c.Close()
		fmt.Println("CLOSED:", fmt.Sprintf("%s:%d", host, port))
		os.Exit(0)
	}()

	fmt.Println("CONNECTED:", fmt.Sprintf("%s:%d", host, port))

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
			lock := cleanedText[1]
			err := c.Acquire(lock)
			if err != nil {
				fmt.Println("Encountered error:", err)
				sigChan <- syscall.SIGINT
				return
			}
		case "release":
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

func getClient() (client.Client, error) {
	c := client.NewClient(&client.ClientOptions{
		Host:       host,
		Port:       uint16(port),
		OnAcquired: func(lock string) { fmt.Print("Acquired ", lock, "\n> ") },
	})
	if err := c.Connect(); err != nil {
		return nil, err
	}

	return c, nil
}
