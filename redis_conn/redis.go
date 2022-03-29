package redis_conn

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
)

func NewRedisConn() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	pong := rdb.Ping(context.Background())

	//look terrible too, but how to check redis availability?
	if pong.Val() != "PONG" {
		log.Fatal("Unable to connect to redis")
	}

	return rdb
}
