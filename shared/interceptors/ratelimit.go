package interceptors

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// RateLimitInterceptor membatasi jumlah request per IP per window waktu.
// Menggunakan Redis Sliding Window agar konsisten di semua replika service.
// Jika Redis down, interceptor akan fail-open (meneruskan request).
func RateLimitInterceptor(redisClient *redis.Client, maxRequests int, window time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Ambil IP klien dari gRPC peer info
		clientIP := "unknown"
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		// Key Redis: ratelimit:<IP>:<method>
		key := fmt.Sprintf("ratelimit:%s:%s", clientIP, info.FullMethod)

		// Increment counter di Redis (auto-expire setelah window)
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			// Jika Redis down, lewatkan rate limiter (fail-open strategy)
			// Lebih baik meneruskan request daripada menolak semua
			return handler(ctx, req)
		}

		// Set TTL hanya saat pertama kali key dibuat (count == 1)
		if count == 1 {
			redisClient.Expire(ctx, key, window)
		}

		// Tolak jika melebihi batas
		if count > int64(maxRequests) {
			return nil, status.Errorf(codes.ResourceExhausted,
				"Rate limit exceeded: maksimum %d request per %s. Silakan coba lagi nanti.",
				maxRequests, window)
		}

		return handler(ctx, req)
	}
}
