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

type user struct {
	ip        string
	userAgent string
	id        uint64
}

type subscriber struct {
	messages  chan []byte
	closeSlow func()
}

type chatServer struct {
	users                  []user
	subscribeMessageBuffer int
	serveMux               http.ServeMux
	subscribers            map[*subscriber]struct{}
}

type response struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func (*chatServer) subscribeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
}

func createServer() *chatServer {
	cs := chatServer{
		users:                  []user{},
		subscribeMessageBuffer: 16,
		serveMux:               *http.NewServeMux(),
		subscribers:            make(map[*subscriber]struct{}),
	}

	cs.serveMux.Handle("/", cs.subscribeHandler())

	return &cs
}
