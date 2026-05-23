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

	var (
		backendConn net.Conn
		err         error
	)
	if count%2 == 0 {
		backendConn, err = net.Dial("tcp", "localhost:3000")
		if err != nil {
			fmt.Println("connection refused, trying backend 2")
			backendConn, err = net.Dial("tcp", "localhost:3001")
			if err != nil {
				fmt.Println("all backends down")
				return
			}
		}
	} else {
		backendConn, err = net.Dial("tcp", "localhost:3001")
		if err != nil {
			fmt.Println("connection 2 refused, trying backend 1")
			backendConn, err = net.Dial("tcp", "localhost:3000")
			if err != nil {
				fmt.Println("all backends down")
				return
			}
		}
	}
	defer backendConn.Close()
	go io.Copy(backendConn, clientConn)
	io.Copy(clientConn, backendConn)
}
