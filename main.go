package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

var (
	count = 0
	mu    sync.Mutex
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleConn(conn)
	}
}

func handleConn(clientConn net.Conn) {
	defer clientConn.Close()
	// Stage 1: connect to backend and forward bytes both ways

	mu.Lock()
	count = count + 1
	mu.Unlock()

	if count%2 == 0 {
		backendConn1, err := net.Dial("tcp", "localhost:3000")
		if err != nil {
			fmt.Println("connection refused")
			return

		}
		defer backendConn1.Close()
		go io.Copy(backendConn1, clientConn)
		io.Copy(clientConn, backendConn1)

	} else {
		backendConn2, err2 := net.Dial("tcp", "localhost:3001")
		if err2 != nil {
			fmt.Println("connection 2 refused")
			return
		}
		defer backendConn2.Close()

		go io.Copy(backendConn2, clientConn)
		io.Copy(clientConn, backendConn2)

	}
}
