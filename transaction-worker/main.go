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

	"transaction-worker/config"
	"transaction-worker/models"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const TransferTopic = "transfer-requests"
const ConsumerGroup  = "transfer-workers"

// TransferMessage adalah representasi pesan dari Kafka
type TransferMessage struct {
	ReferenceNumber string    `json:"reference_number"`
	FromAccountID   string    `json:"from_account_id"`
	ToAccountID     string    `json:"to_account_id"`
	Amount          float64   `json:"amount"`
	SubmittedAt     string    `json:"submitted_at"`
}

func main() {
	config.ConnectDatabase()
	config.ConnectRedis()

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092"
	}

	// Buat Kafka Reader (Consumer) dengan Consumer Group
	// Consumer Group memastikan jika ada lebih dari 1 worker, mereka TIDAK memproses pesan yang sama
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaBroker},
		Topic:          TransferTopic,
		GroupID:        ConsumerGroup,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
	})
	defer reader.Close()

	log.Printf("[transaction-worker] Siap memproses antrean dari topic '%s' (group: %s)\n", TransferTopic, ConsumerGroup)

	// Tangkap sinyal untuk graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("[transaction-worker] Menerima sinyal berhenti, graceful shutdown...")
		cancel()
	}()

	// Loop tak terbatas: terus baca dari Kafka dan proses satu per satu
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break // Context dibatalkan (shutdown)
			}
			log.Printf("[transaction-worker] Error membaca dari Kafka: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("[transaction-worker] Memproses pesan [offset:%d] ref: %s\n", msg.Offset, string(msg.Key))
		processTransfer(msg.Value)
	}

	log.Println("[transaction-worker] Worker berhenti dengan bersih.")
}

// processTransfer melakukan eksekusi ACID ke PostgreSQL Master
// Fungsi ini berjalan secara berurutan (sequential) per pesan, terkendali, tidak berebutan
func processTransfer(messageBytes []byte) {
	var msg TransferMessage
	if err := json.Unmarshal(messageBytes, &msg); err != nil {
		log.Printf("[transaction-worker] ❌ Gagal parse pesan JSON: %v\n", err)
		return
	}

	fromID, err := uuid.Parse(msg.FromAccountID)
	if err != nil {
		log.Printf("[transaction-worker] ❌ Invalid FromAccountID: %v\n", err)
		return
	}
	toID, err := uuid.Parse(msg.ToAccountID)
	if err != nil {
		log.Printf("[transaction-worker] ❌ Invalid ToAccountID: %v\n", err)
		return
	}

	// Eksekusi ACID Transaction ke Master DB
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		var fromAccount models.Account
		var toAccount models.Account

		// Pessimistic Locking - satu per satu, rapi, tidak deadlock
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("account_id = ?", fromID).First(&fromAccount).Error; err != nil {
			return fmt.Errorf("sender_not_found: %v", err)
		}
		if fromAccount.Balance < msg.Amount {
			return fmt.Errorf("insufficient_balance (balance: %.2f, requested: %.2f)", fromAccount.Balance, msg.Amount)
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("account_id = ?", toID).First(&toAccount).Error; err != nil {
			return fmt.Errorf("receiver_not_found: %v", err)
		}

		fromAccount.Balance -= msg.Amount
		toAccount.Balance += msg.Amount

		if err := tx.Save(&fromAccount).Error; err != nil { return err }
		if err := tx.Save(&toAccount).Error; err != nil { return err }

		transactionLog := models.Transaction{
			ReferenceNumber: msg.ReferenceNumber,
			FromAccountID:   fromID,
			ToAccountID:     toID,
			Amount:          msg.Amount,
			TransactionType: "TRANSFER",
			Status:          "SUCCESS",
		}
		if err := tx.Create(&transactionLog).Error; err != nil { return err }

		return nil
	})

	if err != nil {
		log.Printf("[transaction-worker] ❌ GAGAL memproses %s: %v\n", msg.ReferenceNumber, err)
		return
	}

	// Invalidasi Redis Cache setelah transfer berhasil
	config.RedisClient.Del(config.Ctx, fmt.Sprintf("account:balance:%s", fromID.String()))
	config.RedisClient.Del(config.Ctx, fmt.Sprintf("account:balance:%s", toID.String()))

	log.Printf("[transaction-worker] ✅ SUKSES memproses %s (%.2f dari %s ke %s)\n",
		msg.ReferenceNumber, msg.Amount, msg.FromAccountID, msg.ToAccountID)
}
