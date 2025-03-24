package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/coder/websocket"
)

const FALLBACK_PORT string = "8080"

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

func getAddress() string {
	var port string

	if len(os.Args) <= 1 {
		port = FALLBACK_PORT
	} else {
		port = os.Args[1]
	}

	return "ws://localhost:" + port
}

func (cs *chatServer) publishHandler() http.Handler {
	address := getAddress()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{address}})

		if err != nil {
			fmt.Printf("An error occurred: %v\n", err)
		}

		for sub := range cs.subscribers {
			select {
			case <-sub.messages:
			default:
				fmt.Printf("No subscribers to publish to.")
			}
		}
	})
}

func (cs *chatServer) subscribeHandler() http.Handler {
	sub := subscriber{
		messages:  make(chan []byte, cs.subscribeMessageBuffer),
		closeSlow: nil,
		// TODO: handle slow connection
	}

	cs.subscribers[&sub] = struct{}{}
	address := getAddress()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{address}})

		if err != nil {
			fmt.Printf("An error occured: %v\n", err)
		}

		ctx, cancel := context.WithCancel(r.Context())

		defer cancel()

		ctx = c.CloseRead(ctx)

		for {
			select {
			case msg := <-sub.messages:
				fmt.Printf("received message: %v", msg)
				answerMessage(string(msg))
			case <-ctx.Done():
				c.Close(websocket.StatusNormalClosure, "")
			default:
				if ctx.Err() != nil {
					fmt.Print(ctx.Err().Error())
					c.Close(websocket.StatusAbnormalClosure, "")
				}
			}
		}
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
	cs.serveMux.Handle("/publish", cs.publishHandler())

	return &cs
}

func answerMessage(message string) string {
	if message == "Connected to Websocket." {
		return "Successfully connected, hello!"
	}

	return "Received! hello form server."
}
