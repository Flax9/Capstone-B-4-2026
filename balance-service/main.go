package main

import (
	"balance-service/config"
	"balance-service/handlers"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	pb "capstone/proto/balance"
	"shared/interceptors"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	config.ConnectDatabase()
	config.ConnectRedis()

	// 1. Jalankan Prometheus Metrics Server (Background)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Prometheus metrics tersedia di :2112/metrics")
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Printf("Gagal menjalankan metrics server: %v", err)
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "9002"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Gagal membuka port %s: %v", port, err)
	}

	// 2. gRPC Server dengan Chain Interceptor: RateLimit → CircuitBreaker → Tracing → Prometheus
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RateLimitInterceptor(config.RedisClient, 2000, 1*time.Minute),
			interceptors.CircuitBreakerInterceptor("balance-service"),
			interceptors.TracingInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	pb.RegisterBalanceServiceServer(grpcServer, &handlers.BalanceServer{})

	// Register all services to Prometheus
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(grpcServer)

	log.Printf("Balance Service (gRPC) berjalan di port %s [Rate Limit: 2000 req/min]", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Gagal menjalankan server gRPC: %v", err)
	}
}
