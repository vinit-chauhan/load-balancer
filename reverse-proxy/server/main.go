package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	name := os.Getenv("SERVER_NAME")
	if name == "" {
		panic("SERVER_NAME is not set")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("received request from: ", r.RemoteAddr)
		w.Write(([]byte)("Hello from server: " + name))
	})

	fmt.Println("Server is running on port 80")
	http.ListenAndServe(":80", nil)
}
