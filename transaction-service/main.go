package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"
	"transaction-service/config"
	"transaction-service/handlers"

	pb "capstone/proto/transaction"
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
		port = "9003"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Gagal membuka port %s: %v", port, err)
	}

	// 2. gRPC Server dengan Chain Interceptor: RateLimit → CircuitBreaker → Tracing → Prometheus
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RateLimitInterceptor(config.RedisClient, 500, 1*time.Minute),
			interceptors.CircuitBreakerInterceptor("transaction-service"),
			interceptors.TracingInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	pb.RegisterTransactionServiceServer(grpcServer, &handlers.TransactionServer{})

	// Register all services to Prometheus
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(grpcServer)

	log.Printf("Transaction Service (gRPC) berjalan di port %s [Rate Limit: 500 req/min]", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Gagal menjalankan server gRPC: %v", err)
	}
}
