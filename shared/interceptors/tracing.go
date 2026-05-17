package interceptors

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// TraceIDKey adalah key metadata untuk propagasi Trace ID antar-service.
const TraceIDKey = "x-trace-id"

// TracingInterceptor menambahkan Trace ID unik ke setiap request gRPC.
//   - Jika client sudah mengirim Trace ID via metadata, gunakan ID tersebut (propagasi).
//   - Jika belum ada, generate UUID baru.
//   - Setiap request masuk dan keluar akan di-log dengan Trace ID untuk korelasi.
//
// Contoh output log:
//
//	[TRACE:a1b2c3d4-...] → /auth.AuthService/Login (mulai)
//	[TRACE:a1b2c3d4-...] ✓ /auth.AuthService/Login (sukses) [12.34ms]
func TracingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		traceID := ""

		// Cek apakah client sudah mengirim trace ID via gRPC metadata
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get(TraceIDKey); len(values) > 0 {
				traceID = values[0]
			}
		}

		// Generate UUID baru jika belum ada
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Log request masuk
		start := time.Now()
		log.Printf("[TRACE:%s] → %s", traceID, info.FullMethod)

		// Eksekusi handler bisnis
		resp, err := handler(ctx, req)

		// Log request selesai dengan durasi
		duration := time.Since(start)
		if err != nil {
			log.Printf("[TRACE:%s] ✗ %s (error: %v) [%s]", traceID, info.FullMethod, err, duration)
		} else {
			log.Printf("[TRACE:%s] ✓ %s [%s]", traceID, info.FullMethod, duration)
		}

		return resp, err
	}
}
