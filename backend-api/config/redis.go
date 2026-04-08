package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background()
)

func ConnectRedis() {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost" // Standard Dev Fallback
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: "", // No password for dev redis
		DB:       0,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Gagal terhubung ke Redis Cache: %v", err)
	}

	log.Println("Berhasil terhubung ke Redis Layer-2!")
}
