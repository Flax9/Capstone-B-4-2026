package handlers

import (
	"banking-backend/config"
	"banking-backend/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func GetAccountBalance(c *fiber.Ctx) error {
	accountIDStr := c.Params("id")
	
	// Validasi Parameter UUID
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid Account ID format"})
	}

	// 1. Cek Ketersediaan dari Redis Cache (Cache-Aside)
	cacheKey := fmt.Sprintf("account:balance:%s", accountID.String())
	cachedData, err := config.RedisClient.Get(config.Ctx, cacheKey).Result()
	
	if err == nil && cachedData != "" {
		var account models.Account
		_ = json.Unmarshal([]byte(cachedData), &account)
		return c.JSON(fiber.Map{
			"source": "redis_cache", // Pembuktian rute diringankan via Cache
			"data":   account,
		})
	} else if err != redis.Nil {
		fmt.Printf("Redis Get Error: %v\n", err)
	}

	// 2. CACHE MISS => Menyelam ke PostgreSQL (Ber-rute masuk REPLICA)
	var account models.Account
	if result := config.DB.Where("account_id = ?", accountID).First(&account); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
	}

	// 3. Re-Populate (Mengisi Ulang Cache)
	accountJSON, _ := json.Marshal(account)
	_ = config.RedisClient.Set(config.Ctx, cacheKey, accountJSON, 60*time.Second).Err()

	return c.JSON(fiber.Map{
		"source": "postgresql_replica", // Membuktikan terpaksa membentur Database
		"data":   account,
	})
}
