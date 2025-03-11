package main

import (
	"github.com/coder/websocket"
)

type User struct {
	ip        string
	userAgent string
	id        uint64
}

type chatServer struct {
}

type chatClient struct {
	con      websocket.Conn
	messages []Message
	users    []User
}
