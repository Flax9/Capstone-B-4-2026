package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var DB *gorm.DB

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
		log.Fatal("[auth-service] Gagal menyambung ke database Master via PgBouncer: ", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("[auth-service] Gagal mendapatkan instance SQL DB: ", err)
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
		log.Fatal("[auth-service] Gagal memasang DBResolver: ", err)
	}
	log.Println("[auth-service] Terhubung ke PostgreSQL via PgBouncer!")
}
