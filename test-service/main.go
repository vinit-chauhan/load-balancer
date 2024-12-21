package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func main() {
	name := os.Getenv("SERVER_NAME")
	if name == "" {
		panic("SERVER_NAME is not set")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("received request from: ", r.RemoteAddr)
		i, err := w.Write(([]byte)("Hello from server: " + name))
		if err != nil {
			fmt.Println("error sending msg:" + err.Error())
		}
		fmt.Println("sent" + strconv.Itoa(i) + "bytes")
	})

	fmt.Println("Server is running on port 80")
	http.ListenAndServe(":80", nil)
}
