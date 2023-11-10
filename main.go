package main

import (
	"fmt"

	"github.com/maansthoernvik/locksmith/server/connection"
)

func main() {
	fmt.Println("Hello world!")
	connection.Accept()
}
