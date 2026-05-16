package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var DB *gorm.DB
var RedisClient *redis.Client
var Ctx = context.Background()

func ConnectDatabase() {
	masterHost := os.Getenv("DB_MASTER_HOST")
	if masterHost == "" { masterHost = "127.0.0.1" }
	masterPort := os.Getenv("DB_MASTER_PORT")
	if masterPort == "" { masterPort = "6432" }

	replicaHost := os.Getenv("DB_REPLICA_HOST")
	if replicaHost == "" { replicaHost = "127.0.0.1" }
	replicaPort := os.Getenv("DB_REPLICA_PORT")
	if replicaPort == "" { replicaPort = "6432" }

	masterDSN := fmt.Sprintf("host=%s user=user password=password dbname=capstonev2 port=%s sslmode=disable", masterHost, masterPort)
	replicaDSN := fmt.Sprintf("host=%s user=user password=password dbname=capstonev2 port=%s sslmode=disable", replicaHost, replicaPort)

	var err error
	DB, err = gorm.Open(postgres.Open(masterDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("[transaction-service] Gagal menyambung ke Master via PgBouncer: ", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("[transaction-service] Gagal mendapatkan SQL DB: ", err)
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	err = DB.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(masterDSN)},
		Replicas: []gorm.Dialector{postgres.Open(replicaDSN)},
		Policy:   dbresolver.RandomPolicy{},
	}).SetMaxIdleConns(5).SetMaxOpenConns(25).SetConnMaxLifetime(5 * time.Minute))

	if err != nil {
		log.Fatal("[transaction-service] Gagal memasang DBResolver: ", err)
	}
	log.Println("[transaction-service] Terhubung ke PostgreSQL via PgBouncer!")
}

func ConnectRedis() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" { redisHost = "127.0.0.1" }
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" { redisPort = "6379" }

	RedisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})
	if err := RedisClient.Ping(Ctx).Err(); err != nil {
		log.Fatal("[transaction-service] Gagal menyambung ke Redis: ", err)
	}
	log.Println("[transaction-service] Terhubung ke Redis!")
}
