package main

import (
	"banking-backend/config"
	"fmt"
)

func main() {
	fmt.Println("Memulai inisiasi Banking Backend dengan Golang...")

	// Jalankan fungsi koneksi database & Read-Write resolver
	config.ConnectDatabase()

	// Kode migrasi tabel atau server bisa diletakkan di sini...
    // contoh: config.DB.AutoMigrate(&User{}, &Account{})
    
    fmt.Println("Aplikasi berjalan!")
}
