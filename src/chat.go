package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const FALLBACK_PORT string = "8080"

var ORIGIN_PATTERNS = []string{"localhost:1234", "127.0.0.1:1234"}

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
	subscriberMutex        sync.Mutex
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

func startConnection() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1234")
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: ORIGIN_PATTERNS})

		if err != nil {
			fmt.Printf("Error while setting up connection: %v\n", err)
		}

		msg := "Connected to WebSocket"
		ctx, cancel := context.WithCancel(r.Context())

		defer cancel()

		conn.Write(ctx, websocket.MessageText, []byte(msg))
	})
}

func (cs *chatServer) publishHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1234")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		// w.WriteHeader(http.StatusOK)
		_, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: ORIGIN_PATTERNS})

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

func (cs *chatServer) addSubscriber(s *subscriber) {
	cs.subscriberMutex.Lock()
	cs.subscribers[s] = struct{}{}
	cs.subscriberMutex.Unlock()
}

func (cs *chatServer) deleteSubscriber(s *subscriber) {
	cs.subscriberMutex.Lock()
	delete(cs.subscribers, s)
	cs.subscriberMutex.Unlock()
}

func (cs *chatServer) subscribeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1234")
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: ORIGIN_PATTERNS})

		if err != nil {
			fmt.Printf("An error occured: %v\n", err)
		}

		defer c.Close(websocket.StatusNormalClosure, "Connection closed")

		sub := subscriber{
			messages:  make(chan []byte, cs.subscribeMessageBuffer),
			closeSlow: nil,
			// TODO: handle slow connection
		}

		cs.addSubscriber(&sub)
		defer cs.deleteSubscriber(&sub)

		ctx, cancel := context.WithCancel(r.Context())

		defer cancel()

		ctx = c.CloseRead(ctx)

		for {
			select {
			case msg := <-sub.messages:
				fmt.Printf("received message: %v", msg)
				answerMessage(c, ctx)
			case <-ctx.Done():
				return
			}
		}
	})
}

func (cs *chatServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cs.serveMux.ServeHTTP(w, r)
}

func createServer() *chatServer {
	cs := chatServer{
		users:                  []user{},
		subscribeMessageBuffer: 16,
		serveMux:               *http.NewServeMux(),
		subscribers:            make(map[*subscriber]struct{}),
	}

	cs.serveMux.Handle("/", startConnection())
	cs.serveMux.Handle("/subscribe", cs.subscribeHandler())
	cs.serveMux.Handle("/publish", cs.publishHandler())

	return &cs
}

func answerMessage(conn *websocket.Conn, ctx context.Context) {
	ans := "Received! hello from server."

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()

	conn.Write(ctx, websocket.MessageText, []byte(ans))
}
