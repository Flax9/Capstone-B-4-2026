package main

import (
	"log"
	"time"

	"auth-service/config"
	"auth-service/handlers"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	config.ConnectDatabase()
	config.ConnectKafka()

	app := fiber.New(fiber.Config{
		AppName: "Auth Service - Banking Capstone",
		Prefork: true, // Membagi beban ke semua core CPU secara efisien
	})

	prometheus := fiberprometheus.New("auth-service")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	app.Use(logger.New(logger.Config{
		Format:     "[auth-svc] [${time}] ${status} - ${latency} | ${method} ${path}\n",
		TimeFormat: "15:04:05",
	}))

	api := app.Group("/api")
	api.Post("/auth/login", handlers.Login)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("auth-service: OK")
	})

	log.Println("[auth-service] Menyala di port :9001 🔑")
	if err := app.Listen(":9001"); err != nil {
		log.Fatalf("[auth-service] Error: %v", err)
	}
}
