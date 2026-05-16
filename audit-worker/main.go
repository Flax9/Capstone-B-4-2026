package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AuditLog struct {
	LogID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	Action    string    `gorm:"not null"`
	IPAddress string
	UserAgent string
	Details   string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func main() {
	// Connect to DB Master
	dsn := fmt.Sprintf("host=%s user=user password=password dbname=capstonev2 port=%s sslmode=disable", 
		os.Getenv("DB_MASTER_HOST"), os.Getenv("DB_MASTER_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi ke DB:", err)
	}

	// Kafka Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER")},
		Topic:   "audit-logs",
		GroupID: "audit-worker-group",
	})
	defer reader.Close()

	log.Println("[audit-worker] Standby memproses audit logs...")

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			break
		}

		var data map[string]interface{}
		json.Unmarshal(msg.Value, &data)

		userID, _ := uuid.Parse(fmt.Sprintf("%v", data["user_id"]))
		
		logEntry := AuditLog{
			UserID:    userID,
			Action:    fmt.Sprintf("%v", data["action"]),
			IPAddress: fmt.Sprintf("%v", data["ip_address"]),
			UserAgent: fmt.Sprintf("%v", data["user_agent"]),
			Details:   fmt.Sprintf("%v", data["details"]),
		}

		if err := db.Create(&logEntry).Error; err != nil {
			log.Println("Gagal simpan audit log:", err)
		}
	}
}
