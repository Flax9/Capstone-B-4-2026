package main

import (
	"log"
	"time"

	"transaction-service/config"
	"transaction-service/handlers"
	"transaction-service/middlewares"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	config.ConnectDatabase()
	config.ConnectRedis()
	config.ConnectKafka()

	app := fiber.New(fiber.Config{
		AppName: "Transaction Service - Banking Capstone",
		Prefork: true,
	})

	prometheus := fiberprometheus.New("transaction-service")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	app.Use(logger.New(logger.Config{
		Format:     "[txn-svc] [${time}] ${status} - ${latency} | ${method} ${path}\n",
		TimeFormat: "15:04:05",
	}))

	api := app.Group("/api")
	api.Use(middlewares.Protected())
	api.Post("/transactions/transfer", handlers.Transfer)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("transaction-service: OK")
	})

	log.Println("[transaction-service] Menyala di port :9003 💸")
	if err := app.Listen(":9003"); err != nil {
		log.Fatalf("[transaction-service] Error: %v", err)
	}
}
