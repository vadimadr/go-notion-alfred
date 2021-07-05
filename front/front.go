package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	start := time.Now()
	log.SetOutput(os.Stderr)
	defer log.Println("Got response in ", time.Now().Sub(start))

	query := os.Args[1]
	log.Printf("Query: \"%s\"", query)

	// connect to socket
	c, err := net.Dial("unix", "/tmp/tmp.socket")
	HandleError(err)
	defer c.Close()

	// send query
	_, err = c.Write([]byte(query))
	HandleError(err)

	// read response
	buf := make([]byte, 4096)
	nr, err := c.Read(buf)
	HandleError(err)

	fmt.Print(string(buf[0:nr]))
}

func HandleError(err error) {
	if err == nil {
		return
	}
	log.Fatal(err)
}
