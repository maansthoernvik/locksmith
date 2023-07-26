package main

import (
	"fmt"

	"github.com/maansthoernvik/locksmith/lib/connection"
)

func main() {
	fmt.Println("I'm a client")
	connection.Example()
}
