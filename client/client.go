package main

import (
	"log"
	"net"
	"time"
)

// type something struct {
// 	someValue string
// }

// var globalMap = make(map[string]something, 2)

func main() {
	conn, err := net.Dial("tcp", "localhost:9000")
	log.Println("Connected to localhost:9000", err)

	log.Println("Writing...")

	// log.Println("initial map: \n", globalMap)
	// globalMap["123"] = something{someValue: "123"}
	// log.Println("added one field: \n", globalMap)
	// smthng := &something{someValue: "abc"}
	// globalMap["abc"] = *smthng
	// log.Println("added another field: \n", globalMap)
	// smthng.someValue = "CHANGED!"
	// log.Println("changed local struct without updating map: \n", globalMap)

	// return

	//nolint
	conn.Write(Acquire())

	// await acquisition notification...
	bytes := make([]byte, 10)
	n, err := conn.Read(bytes)
	if err != nil {
		log.Fatalln("Failed to read bytes:", err)
	}
	log.Println("Got", n, "bytes:", bytes)

	//nolint
	conn.Write(Release())

	time.Sleep(1 * time.Second)

	conn.Close()
}

func Acquire() []byte {
	return []byte{
		0x0, 0x2, 0x48, 0x48,
	}
}

func Release() []byte {
	return []byte{
		0x1, 0x2, 0x48, 0x48,
	}
}
