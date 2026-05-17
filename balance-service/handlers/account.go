package handlers

import (
	"balance-service/config"
	"balance-service/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "capstone/proto/balance"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type BalanceServer struct {
	pb.UnimplementedBalanceServiceServer
}

func (s *BalanceServer) CheckBalance(ctx context.Context, req *pb.BalanceRequest) (*pb.BalanceResponse, error) {
	accountIDStr := req.UserId // Di asumsi param adalah account id
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return &pb.BalanceResponse{
			StatusCode: 400,
			Message:    "Invalid Account ID format",
		}, nil
	}

	// 1. Cek Redis Cache (Cache-Aside)
	cacheKey := fmt.Sprintf("account:balance:%s", accountID.String())
	cachedData, err := config.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		var account models.Account
		_ = json.Unmarshal([]byte(cachedData), &account)
		return &pb.BalanceResponse{
			StatusCode:     200,
			Message:        "Success (Cache)",
			CurrentBalance: account.Balance,
			AccountNumber:  account.AccountNumber,
		}, nil
	} else if err != redis.Nil {
		fmt.Printf("[balance-service] Redis Get Error: %v\n", err)
	}

	// 2. CACHE MISS => Query Replica PostgreSQL
	var account models.Account
	if result := config.DB.WithContext(ctx).Where("account_id = ?", accountID).First(&account); result.Error != nil {
		return &pb.BalanceResponse{
			StatusCode: 404,
			Message:    "Account not found",
		}, nil
	}

	// 3. Populate Redis Cache
	accountJSON, _ := json.Marshal(account)
	_ = config.RedisClient.Set(ctx, cacheKey, accountJSON, 60*time.Second).Err()

	return &pb.BalanceResponse{
		StatusCode:     200,
		Message:        "Success (DB Replica)",
		CurrentBalance: account.Balance,
		AccountNumber:  account.AccountNumber,
	}, nil
}
