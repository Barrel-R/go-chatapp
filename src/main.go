package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting program")
	// createRedisClient()
	createServer()

	go func() {
		err := http.ListenAndServe("localhost:8080", nil)

		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	select {} // keep the program running
}
