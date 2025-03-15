package main

import (
	"time"
)

type User struct {
	ip        string
	userAgent string
	id        uint64
}

type chatServer struct {
	users []User
}

type response struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
