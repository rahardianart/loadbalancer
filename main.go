package main

import (
	"fmt"
	"io"
	"net"
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

	backendConn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		fmt.Println("connection refused")
		return
	}

	defer backendConn.Close()

	go io.Copy(backendConn, clientConn)
	io.Copy(clientConn, backendConn)
}
