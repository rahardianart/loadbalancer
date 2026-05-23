package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := "3000"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("backend:%s received %s %s", port, r.Method, r.URL.Path)
		fmt.Fprintf(w, "hello from backend on port %s\n", port)
	})

	log.Printf("backend listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
