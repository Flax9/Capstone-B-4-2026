package main

import (
	"log"

	"banking-backend/config"
	"banking-backend/handlers"
	"banking-backend/middlewares"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/ansrivas/fiberprometheus/v2"
)

func main() {
	// 1. Inisialisasi Koneksi ke Infrastruktur Docker
	config.ConnectDatabase()
	config.ConnectRedis()

	// 2. Inisiasi Framework Web Tercepat 'Fiber'
	app := fiber.New(fiber.Config{
		AppName: "Banking MVP Capstone (Fiber GORM Redis)",
	})

	// ==============================================
	// 🧿 2A. SENSOR PROMETHEUS + R.E.D METRICS (TAHAP 4) 🧿
	// Menghubungkan middleware Fiber dengan Prometheus untuk menangkap metrik API
	// (Requests per second, Error Rate, dan HTTP Duration/Latency)
	// ==============================================
	prometheus := fiberprometheus.New("spring-boot-backend")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// 3A. Menambahkan Terminal Logger agar request terlihat cantik di konsol
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} | ${method} ${path}\n",
		TimeFormat: "15:04:05",
	}))

	// 3B. Memasang Baju Zirah Pembatasan Kecepatan (Rate Limiter Anti-Spam / DDoS)
	// Hanya mengizinkan maksimal 60 Panggilan HTTP selama rentang waktu 1 Menit dari IP yang sama.
	app.Use(limiter.New(limiter.Config{
		// Celah Pelolosan Otomatis untuk Mesin K6 Rafael (SRE Load Tester)
		Next: func(c *fiber.Ctx) bool {
			return c.Get("X-Test-Bypass") == "b7fc809a-super-secret-key-capstone"
		},
		Max:        60,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Serangan L7 Terdeteksi: Lalu lintas IP Anda melampaui ambang batas 60 request/menit. Anda diblokir sementara.",
			})
		},
	}))

	// 4. Mendaftarkan Grup Rute-Rute MVP Perbankan
	api := app.Group("/api")

	// ==============================================
	// 4A. Rute Publik (Tanpa Perisai JWT)
	// ==============================================
	api.Post("/auth/login", handlers.Login)

	// ==============================================
	// 🛡️ DINDING PENJAGA MIDDLEWARE (HS256) 🛡️
	// Semua rute di bawah baris ini WAJIB membawa Token JWT Valid!
	// ==============================================
	api.Use(middlewares.Protected())

	// ---> Endpoint Beban Berat (Menuju Master Database + Menyetrum Invalidate Redis)
	api.Post("/transactions/transfer", handlers.Transfer)

	// ---> Endpoint Beban Ringan (Cegatan Awal Redis Cache / Menempel Replica DB)
	api.Get("/accounts/:id", handlers.GetAccountBalance)

	// ---> Rute Pengecekan Detak Jantung Kubernetes/Docker
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Bypass Master-Replica Infrastructure: OK v1.0")
	})

	log.Println("Server API menyala agresif pada port :9000 🚀")
	if err := app.Listen(":9000"); err != nil {
		log.Fatalf("Server Fiber tersandung saat boot: %v", err)
	}
}
