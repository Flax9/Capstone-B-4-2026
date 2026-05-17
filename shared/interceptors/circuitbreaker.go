package interceptors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sony/gobreaker/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// Gauge: status circuit breaker saat ini (0=Closed, 1=Half-Open, 2=Open)
	circuitBreakerState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "circuit_breaker_state",
		Help: "Status Circuit Breaker: 0=Closed, 1=Half-Open, 2=Open",
	}, []string{"service"})

	// Counter: total request yang ditolak oleh circuit breaker
	circuitBreakerRejections = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "circuit_breaker_rejections_total",
		Help: "Total request yang ditolak karena circuit breaker terbuka",
	}, []string{"service"})

	// Counter: total perubahan state circuit breaker
	circuitBreakerStateChanges = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "circuit_breaker_state_changes_total",
		Help: "Total perubahan state circuit breaker",
	}, []string{"service", "from", "to"})
)

// CircuitBreakerInterceptor melindungi service dari cascading failure.
// Jika lebih dari 50% request gagal dalam interval 30 detik, circuit akan terbuka
// dan langsung menolak semua request baru selama 10 detik (tanpa memukul database).
func CircuitBreakerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	// Set initial state ke Closed (0)
	circuitBreakerState.WithLabelValues(serviceName).Set(0)

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

		// OnStateChange: Log + update metrik Prometheus setiap perubahan state
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("[CIRCUIT BREAKER] %s: %s → %s", name, from.String(), to.String())

			// Update Prometheus gauge (0=Closed, 1=HalfOpen, 2=Open)
			stateValue := float64(0)
			switch to {
			case gobreaker.StateHalfOpen:
				stateValue = 1
			case gobreaker.StateOpen:
				stateValue = 2
			}
			circuitBreakerState.WithLabelValues(serviceName).Set(stateValue)
			circuitBreakerStateChanges.WithLabelValues(serviceName, from.String(), to.String()).Inc()
		},
	})

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		result, err := cb.Execute(func() (interface{}, error) {
			return handler(ctx, req)
		})

		if err != nil {
			// Jika circuit sedang terbuka, kembalikan error Unavailable
			if err == gobreaker.ErrOpenState {
				circuitBreakerRejections.WithLabelValues(serviceName).Inc()
				return nil, status.Errorf(codes.Unavailable,
					"Service %s sedang tidak tersedia (circuit breaker OPEN). Silakan coba lagi dalam beberapa detik.",
					serviceName)
			}
			if err == gobreaker.ErrTooManyRequests {
				circuitBreakerRejections.WithLabelValues(serviceName).Inc()
				return nil, status.Errorf(codes.Unavailable,
					"Service %s sedang dalam pemulihan (circuit breaker HALF-OPEN). Silakan tunggu.",
					serviceName)
			}
			return nil, err
		}

		return result, nil
	}
}
