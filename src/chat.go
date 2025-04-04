package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/redis/go-redis/v9"
)

const FALLBACK_PORT string = "8080"

var ORIGIN_PATTERNS = []string{"localhost:1234"}

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
	rdb                    *redis.Client
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight (OPTIONS) requests
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1234")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent) // 204 No Content for preflight requests
			return
		}

		// Set CORS headers for actual requests
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1234")

		body := http.MaxBytesReader(w, r.Body, 8192)
		msg, err := io.ReadAll(body)

		if err != nil {
			fmt.Printf("Error while reading body: %v\n", err)
			http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		}

		cs.subscriberMutex.Lock()
		defer cs.subscriberMutex.Unlock()

		fmt.Printf("sub count: %v\n", len(cs.subscribers))

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = cs.rdb.Publish(ctx, "subscribers", msg).Err()

		if err != nil {
			fmt.Printf("Error publishing message: %v\n", err)
		}

		w.WriteHeader(http.StatusAccepted)
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
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: ORIGIN_PATTERNS, InsecureSkipVerify: true,
		})

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

		pubsub := cs.rdb.Subscribe(ctx, "subscribers")

		for {
			msg, err := pubsub.ReceiveMessage(ctx)

			if err != nil {
				fmt.Printf("Error when receiving message: %v\n", err)
			}

			fmt.Printf("Received: %v\n", msg.Payload)

			answerMessage(c, ctx)
		}
	})
}

func answerMessage(conn *websocket.Conn, ctx context.Context) {
	ans := "Received! hello from server."

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()

	conn.Write(ctx, websocket.MessageText, []byte(ans))
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
		rdb:                    createRedisClient(),
	}

	cs.serveMux.Handle("/subscribe", cs.subscribeHandler())
	cs.serveMux.Handle("/publish", cs.publishHandler())

	return &cs
}
