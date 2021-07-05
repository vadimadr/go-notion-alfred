package main

import (
	"net"
	"os"
)

func main() {
	// connect to socket
	os.Remove("/tmp/tmp.socket")
	l, err := net.Listen("unix", "/tmp/tmp.socket")
	HandleError(err)
	defer l.Close()

	worker := CreateWorker()

	// wait for connections
	for {
		conn, err := l.Accept()
		HandleError(err)
		go ServeConnections(conn, worker)
	}
}

func ServeConnections(c net.Conn, worker *Worker) {
	buf := make([]byte, 10000)
	nr, err := c.Read(buf)
	HandleError(err)

	query := string(buf[:nr])

	resp := worker.Search(query)

	_, err = c.Write(resp)
	HandleError(err)
}

func HandleError(err error) {
	if err == nil {
		return
	}
	// log.Fatal(err)
}
