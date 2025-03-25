package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	err := run()

	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	fmt.Println("Starting program")
	// createRedisClient()

	l, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Printf("Error setting up TCP: %v\n", err)
	}

	cs := createServer()

	s := &http.Server{
		Handler:      cs,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	errc := make(chan error, 1)

	go func() {
		errc <- s.Serve(l)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	select {
	case err := <-errc:
		log.Printf("Failed to serve: %v\n", err)
	case sig := <-sigs:
		log.Printf("terminating: %v\n", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return s.Shutdown(ctx)
}
