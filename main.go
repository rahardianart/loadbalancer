package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	count = 0
	mu    sync.Mutex

	backendAHealth bool
	backendBHealth bool
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {

			aHealth := checkHealth("localhost:3000", 2*time.Second)
			bHealth := checkHealth("localhost:3001", 2*time.Second)

			mu.Lock()
			backendAHealth = aHealth
			backendBHealth = bHealth

			mu.Unlock()
		}
	}()

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
	// Stage 1:ways: connect to backend and forward bytes both ways

	mu.Lock()
	count = count + 1
	mu.Unlock()

	var (
		backendConn net.Conn
		err         error
		primary     string
		fallback    string
	)
	if count%2 == 0 {
		mu.Lock()
		if backendAHealth {
			primary = "localhost:3000"
			fallback = "localhost:3001"

		} else {
			primary = "localhost:3001"
			fallback = "localhost:3000"

		}
		mu.Unlock()
		backendConn, err = net.Dial("tcp", primary)
		if err != nil {
			fmt.Println("connection refused, trying backend 2")

			backendConn, err = net.Dial("tcp", fallback)
			if err != nil {
				return
			}
		}
	} else {
		mu.Lock()
		if backendBHealth {
			primary = "localhost:3001"
			fallback = "localhost:3000"
		} else {
			primary = "localhost:3000"
			fallback = "localhost:3001"
		}
		mu.Unlock()

		backendConn, err = net.Dial("tcp", primary)
		if err != nil {
			fmt.Println("connection 2 refused, trying backend 1")
			backendConn, err = net.Dial("tcp", fallback)
			if err != nil {
				return
			}
		}

	}
	defer backendConn.Close()
	go io.Copy(backendConn, clientConn)
	io.Copy(clientConn, backendConn)
}

func checkHealth(address string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}

	conn.Close()
	return true
}
