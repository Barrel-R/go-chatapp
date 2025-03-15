package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting program")
	// createRedisClient()
	http.Handle("/", createServer())
	go func() {
		err := http.ListenAndServe(":8080", nil)

		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	select {} // keep the program running
}
