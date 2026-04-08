package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var DB *gorm.DB

func ConnectDatabase() {
	masterHost := os.Getenv("DB_MASTER_HOST")
	if masterHost == "" { masterHost = "127.0.0.1" }
	masterPort := os.Getenv("DB_MASTER_PORT")
	if masterPort == "" { masterPort = "15432" }
	
	replicaHost := os.Getenv("DB_REPLICA_HOST")
	if replicaHost == "" { replicaHost = "127.0.0.1" }
	replicaPort := os.Getenv("DB_REPLICA_PORT")
	if replicaPort == "" { replicaPort = "15433" }

	masterDSN := fmt.Sprintf("host=%s user=user password=password dbname=capstonev2 port=%s sslmode=disable", masterHost, masterPort)
	replicaDSN := fmt.Sprintf("host=%s user=user password=password dbname=capstonev2 port=%s sslmode=disable", replicaHost, replicaPort)

	var err error

	// 1. Koneksi awal selalu ke Master (Utama)
	DB, err = gorm.Open(postgres.Open(masterDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal menyambung ke database Master: ", err)
	}

	// 2. Pasang DBResolver untuk routing Read/Write otomatis (Sesuai Pattern Master-Replica)
	err = DB.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(masterDSN)},  // Untuk instruksi INSERT, UPDATE, DELETE
		Replicas: []gorm.Dialector{postgres.Open(replicaDSN)}, // Spesialis instruksi membaca (SELECT)
		Policy:   dbresolver.RandomPolicy{},
	}))

	if err != nil {
		log.Fatal("Gagal memasang DBResolver plugin: ", err)
	}

	log.Println("Berhasil menyambung ke Master & Replica PostgreSQL menggunakan GORM!")
}
