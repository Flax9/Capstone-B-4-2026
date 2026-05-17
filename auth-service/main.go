package main

import (
	"auth-service/config"
	"auth-service/handlers"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	pb "capstone/proto/auth"
	"shared/interceptors"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	config.ConnectDatabase()
	config.ConnectKafka()
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
		port = "9001"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Gagal membuka port %s: %v", port, err)
	}

	// 2. gRPC Server dengan Chain Interceptor: RateLimit → CircuitBreaker → Tracing → Prometheus
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RateLimitInterceptor(config.RedisClient, 1000, 1*time.Minute),
			interceptors.CircuitBreakerInterceptor("auth-service"),
			interceptors.TracingInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	pb.RegisterAuthServiceServer(grpcServer, &handlers.AuthServer{})

	// Register all services to Prometheus
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(grpcServer)

	log.Printf("Auth Service (gRPC) berjalan di port %s [Rate Limit: 1000 req/min]", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Gagal menjalankan server gRPC: %v", err)
	}
}
