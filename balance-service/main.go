package main

import (
	"log"
	"time"

	"balance-service/config"
	"balance-service/handlers"
	"balance-service/middlewares"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	config.ConnectDatabase()
	config.ConnectRedis()

	app := fiber.New(fiber.Config{
		AppName: "Balance Service - Banking Capstone",
		Prefork: true,
	})

	prometheus := fiberprometheus.New("balance-service")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	app.Use(logger.New(logger.Config{
		Format:     "[balance-svc] [${time}] ${status} - ${latency} | ${method} ${path}\n",
		TimeFormat: "15:04:05",
	}))

	api := app.Group("/api")
	api.Use(middlewares.Protected())
	api.Get("/accounts/:id", handlers.GetAccountBalance)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("balance-service: OK")
	})

	log.Println("[balance-service] Menyala di port :9002 💰")
	if err := app.Listen(":9002"); err != nil {
		log.Fatalf("[balance-service] Error: %v", err)
	}
}
