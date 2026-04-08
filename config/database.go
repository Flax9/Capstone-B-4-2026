package config

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var DB *gorm.DB

func ConnectDatabase() {
    // Karena Go menggunakan format DSN bawaan Postgres (bukan JDBC seperti Java),
    // kita sesuaikan formatnya di sini. Anggap ini membaca dari .env Anda.
    
	masterDSN := "host=host.docker.internal user=postgres password=purqon dbname=capstonev2 port=5432 sslmode=disable"
	replicaDSN := "host=host.docker.internal user=postgres password=purqon dbname=capstonev2 port=5432 sslmode=disable"

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
