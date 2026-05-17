package interceptors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sony/gobreaker/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CircuitBreakerInterceptor melindungi service dari cascading failure.
// Jika lebih dari 50% request gagal dalam interval 30 detik, circuit akan terbuka
// dan langsung menolak semua request baru selama 10 detik (tanpa memukul database).
//
// State Machine:
//
//	CLOSED (normal) → OPEN (tolak semua) → HALF-OPEN (coba 5 req) → CLOSED
func CircuitBreakerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	cb := gobreaker.NewCircuitBreaker[interface{}](gobreaker.Settings{
		Name:        fmt.Sprintf("cb-%s", serviceName),
		MaxRequests: 5,                // Izinkan 5 request percobaan saat half-open
		Interval:    30 * time.Second, // Reset counter kegagalan setiap 30 detik
		Timeout:     10 * time.Second, // Coba half-open setelah 10 detik dalam state open

		// ReadyToTrip: Circuit terbuka jika ≥50% request gagal DAN minimal 10 request sudah masuk
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < 10 {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= 0.5
		},

		// OnStateChange: Log setiap perubahan state circuit breaker
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("[CIRCUIT BREAKER] %s: %s → %s", name, from.String(), to.String())
		},
	})

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		result, err := cb.Execute(func() (interface{}, error) {
			return handler(ctx, req)
		})

		if err != nil {
			// Jika circuit sedang terbuka, kembalikan error Unavailable
			if err == gobreaker.ErrOpenState {
				return nil, status.Errorf(codes.Unavailable,
					"Service %s sedang tidak tersedia (circuit breaker OPEN). Silakan coba lagi dalam beberapa detik.",
					serviceName)
			}
			if err == gobreaker.ErrTooManyRequests {
				return nil, status.Errorf(codes.Unavailable,
					"Service %s sedang dalam pemulihan (circuit breaker HALF-OPEN). Silakan tunggu.",
					serviceName)
			}
			return nil, err
		}

		return result, nil
	}
}
