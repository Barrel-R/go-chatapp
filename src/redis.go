package main

import (
	"github.com/redis/go-redis/v9"
)

type Message struct {
	data []byte
}

func createRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	return client
}
