package handlers

import (
	"auth-service/config"
	"auth-service/models"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	pb "capstone/proto/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var user models.User
	var actionEvent string
	var resStatus int32
	var message string
	var tokenString string
	var userFullname string
	var details map[string]interface{}
	var userFound bool

	// 1. CEK REDIS CACHE DULU (Cache-Aside Pattern)
	cacheKey := fmt.Sprintf("user:login:%s", req.Username)
	cachedData, err := config.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		// Cache HIT — tidak perlu query SQL
		_ = json.Unmarshal([]byte(cachedData), &user)
		userFound = true
	} else if err != redis.Nil {
		fmt.Printf("[auth-service] Redis Get Error: %v\n", err)
	}

	// 2. CACHE MISS => Query ke Replica PostgreSQL (timeout 3 detik)
	if !userFound {
		queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		dbErr := config.DB.WithContext(queryCtx).Where("username = ?", req.Username).First(&user).Error
		if dbErr != nil {
			// Bedakan: "record not found" (normal) vs error koneksi DB (fatal)
			if dbErr.Error() == "record not found" {
				userFound = false
			} else {
				// Error koneksi DB / timeout → return gRPC error agar circuit breaker terpicu
				return nil, fmt.Errorf("database error: %v", dbErr)
			}
		} else {
			userFound = true
			// 3. Populate Cache (TTL 60 detik)
			userJSON, _ := json.Marshal(user)
			_ = config.RedisClient.Set(ctx, cacheKey, userJSON, 60*time.Second).Err()
		}
	}

	if !userFound {
		actionEvent = "LOGIN_FAILED_NOTFOUND"
		resStatus = 401
		message = "Identitas gagal divalidasi"
		details = map[string]interface{}{"reason": "user_missing_or_typo"}
	} else {
		claims := jwt.MapClaims{
			"sub":      user.UserID,
			"username": user.Username,
			"exp":      time.Now().Add(time.Minute * 15).Unix(),
			"iat":      time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			secretKey = "capstone_rahasia_negara_2026"
		}

		signedToken, errSigning := token.SignedString([]byte(secretKey))
		if errSigning != nil {
			return &pb.LoginResponse{
				StatusCode: 500,
				Message:    "Gagal menenun tanda tangan kriptografi JWT",
			}, nil
		}

		actionEvent = "LOGIN_SUCCESS"
		resStatus = 200
		message = "Auth Berhasil Disetujui"
		tokenString = signedToken
		userFullname = user.FullName
		details = map[string]interface{}{"auth_method": "password_verification"}
	}

	// Extract IP and User-Agent from gRPC Metadata
	var ipAddress string
	var userAgent string
	if p, ok := peer.FromContext(ctx); ok {
		ipAddress = p.Addr.String()
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if agents := md.Get("user-agent"); len(agents) > 0 {
			userAgent = agents[0]
		}
	}

	// KIRIM AUDIT LOG KE KAFKA
	detailsJSON, _ := json.Marshal(details)
	auditMessage := map[string]interface{}{
		"action":     actionEvent,
		"user_id":    user.UserID,
		"ip_address": ipAddress,
		"user_agent": userAgent,
		"details":    string(detailsJSON),
		"created_at": time.Now().Format(time.RFC3339),
	}
	messageBytes, _ := json.Marshal(auditMessage)

	// Async write to Kafka
	config.KafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(req.Username),
		Value: messageBytes,
	})

	return &pb.LoginResponse{
		StatusCode:    resStatus,
		Message:       message,
		Token:         tokenString,
		UserFullname:  userFullname,
	}, nil
}
