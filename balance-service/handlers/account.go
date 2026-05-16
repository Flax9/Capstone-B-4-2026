package handlers

import (
	"balance-service/config"
	"balance-service/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func GetAccountBalance(c *fiber.Ctx) error {
	accountIDStr := c.Params("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid Account ID format"})
	}

	// 1. Cek Redis Cache (Cache-Aside)
	cacheKey := fmt.Sprintf("account:balance:%s", accountID.String())
	cachedData, err := config.RedisClient.Get(config.Ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		var account models.Account
		_ = json.Unmarshal([]byte(cachedData), &account)
		return c.JSON(fiber.Map{
			"source": "redis_cache",
			"data":   account,
		})
	} else if err != redis.Nil {
		fmt.Printf("[balance-service] Redis Get Error: %v\n", err)
	}

	// 2. CACHE MISS => Query Replica PostgreSQL
	var account models.Account
	if result := config.DB.Where("account_id = ?", accountID).First(&account); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
	}

	// 3. Populate Redis Cache
	accountJSON, _ := json.Marshal(account)
	_ = config.RedisClient.Set(config.Ctx, cacheKey, accountJSON, 60*time.Second).Err()

	return c.JSON(fiber.Map{
		"source": "postgresql_replica",
		"data":   account,
	})
}
