package connection

import (
	"fmt"
	"net"
)

const TCP = "tcp"

func Connect(address string, port uint16) (net.Conn, error) {
	fmt.Println("connection Example function called!")

	conn, err := net.Dial(TCP, fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		fmt.Println("oh noooo")
	} else if err == nil {
		fmt.Println("great!")
	}

	return conn, err
}
