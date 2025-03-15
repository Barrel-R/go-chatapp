package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func createServer() http.HandlerFunc {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"localhost:1234"}})

		if err != nil {
			fmt.Printf("An error occured: %v\n", err)
		}

		defer c.CloseNow()

		ctx, cancel := context.WithCancel(r.Context())

		defer cancel()

		var v interface{}

		for {
			err = wsjson.Read(ctx, c, &v)

			if err != nil {
				fmt.Printf("Error when reading message: %v\n", err)
				break
			}

			var responseMessage string

			if v == "Connected to Websocket." {
				responseMessage = "Successfully connected, hello!"
			} else {
				responseMessage = "Received! hello form server."
			}

			err = wsjson.Write(ctx, c, response{responseMessage, time.Now()})

			log.Printf("Received: %v\n", v)

			if err != nil {
				fmt.Printf("Error while writing back to WebSocket: %v\n", err)
				break
			}
		}

		c.Close(websocket.StatusNormalClosure, "")
	})

	return handler
}

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
