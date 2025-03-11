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
		c, err := websocket.Accept(w, r, nil)

		if err != nil {
			fmt.Printf("An error occured: %v\n", err)
		}

		defer c.CloseNow()

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()

		var v interface{}
		err = wsjson.Read(ctx, c, &v)

		if err != nil {
			fmt.Printf("Error when reading message: %v\n", err)
		}

		log.Printf("Received: %v\n", v)

		c.Close(websocket.StatusNormalClosure, "")
	})

	return handler
}

func createClient() {
	time.Sleep(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, r, err := websocket.Dial(ctx, "ws://localhost:8080", nil)

	if conn == nil {
		fmt.Printf("Couldn't connect, r: %+v\n", r)
	}

	if err != nil {
		fmt.Printf("An error occured: %v\n", err)
	}

	defer conn.CloseNow()

	fmt.Println("Connected to Websocket")

	err = wsjson.Write(ctx, conn, "hello")

	if err != nil {
		fmt.Printf("Error when writing to websocket: %v\n", err)
	}

	conn.Close(websocket.StatusNormalClosure, "")

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

	createClient()

	select {}
}
