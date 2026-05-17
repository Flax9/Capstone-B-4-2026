package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"transaction-worker/config"
	"transaction-worker/models"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

const TransferTopic = "transfer-requests"
const ConsumerGroup = "transfer-workers"

type TransferMessage struct {
	ReferenceNumber string  `json:"reference_number"`
	FromAccountID   string  `json:"from_account_id"`
	ToAccountID     string  `json:"to_account_id"`
	Amount          float64 `json:"amount"`
	SubmittedAt     string  `json:"submitted_at"`
}

// In-Memory State
var (
	accountState = make(map[string]float64)
	stateMutex   sync.RWMutex
	
	// Batch processing
	pendingLogs []models.Transaction
	dirtyAccounts map[string]float64
	batchMutex  sync.Mutex
)

func loadInitialState() {
	log.Println("[LMAX-Engine] Memuat status saldo ke dalam memori (RAM)...")
	var accounts []models.Account
	config.DB.Find(&accounts)
	
	stateMutex.Lock()
	defer stateMutex.Unlock()
	for _, acc := range accounts {
		accountState[acc.AccountID.String()] = acc.Balance
	}
	log.Printf("[LMAX-Engine] %d akun dimuat ke RAM. Mesin siap beroperasi pada kecepatan Nanosecond.", len(accounts))
}

func main() {
	config.ConnectDatabase()
	config.ConnectRedis()

	loadInitialState()
	dirtyAccounts = make(map[string]float64)

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaBroker},
		Topic:          TransferTopic,
		GroupID:        ConsumerGroup,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
	defer reader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("[LMAX-Engine] Graceful shutdown...")
		flushBatchToDB() // flush sisa data
		cancel()
	}()

	// Flusher background task (Asynchronous Snapshotting)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				flushBatchToDB()
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Printf("[LMAX-Engine] Mendengarkan event transfer...\n")

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			time.Sleep(1 * time.Second)
			continue
		}

		// Proses murni di dalam RAM (Sangat Cepat)
		processTransferInMemory(msg.Value)
	}
}

func processTransferInMemory(messageBytes []byte) {
	var msg TransferMessage
	if err := json.Unmarshal(messageBytes, &msg); err != nil {
		return
	}

	stateMutex.Lock()
	defer stateMutex.Unlock()

	fromBal, fromExists := accountState[msg.FromAccountID]
	toBal, toExists := accountState[msg.ToAccountID]

	if !fromExists || !toExists {
		log.Printf("[LMAX-Engine] ❌ Akun tidak ditemukan untuk TRX %s\n", msg.ReferenceNumber)
		return
	}

	if fromBal < msg.Amount {
		// Insufficient balance
		return
	}

	// 1. Mutasi State di RAM
	accountState[msg.FromAccountID] = fromBal - msg.Amount
	accountState[msg.ToAccountID] = toBal + msg.Amount

	// 2. Catat perubahan untuk di-flush nanti
	batchMutex.Lock()
	defer batchMutex.Unlock()
	
	dirtyAccounts[msg.FromAccountID] = accountState[msg.FromAccountID]
	dirtyAccounts[msg.ToAccountID] = accountState[msg.ToAccountID]

	fromID, _ := uuid.Parse(msg.FromAccountID)
	toID, _ := uuid.Parse(msg.ToAccountID)

	pendingLogs = append(pendingLogs, models.Transaction{
		ReferenceNumber: msg.ReferenceNumber,
		FromAccountID:   fromID,
		ToAccountID:     toID,
		Amount:          msg.Amount,
		TransactionType: "TRANSFER",
		Status:          "SUCCESS",
	})
}

// Flush data secara berkala (Batching)
func flushBatchToDB() {
	batchMutex.Lock()
	if len(pendingLogs) == 0 {
		batchMutex.Unlock()
		return
	}

	// Copy data to avoid blocking the main processor
	logsToInsert := pendingLogs
	accountsToUpdate := dirtyAccounts
	
	pendingLogs = make([]models.Transaction, 0)
	dirtyAccounts = make(map[string]float64)
	batchMutex.Unlock()

	// Eksekusi Batch Update ke PostgreSQL
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Update saldos
		for accIDStr, newBalance := range accountsToUpdate {
			accID, _ := uuid.Parse(accIDStr)
			if err := tx.Model(&models.Account{}).Where("account_id = ?", accID).Update("balance", newBalance).Error; err != nil {
				return err
			}
			// Segera update Read Model (Redis)
			cacheKey := fmt.Sprintf("account:balance:%s", accIDStr)
			config.RedisClient.Del(context.Background(), cacheKey) // Invalidasi cepat
		}
		
		// Insert logs in batch
		if len(logsToInsert) > 0 {
			if err := tx.Create(&logsToInsert).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("[LMAX-Engine] ❌ Gagal Flush Batch ke DB: %v\n", err)
	} else {
		log.Printf("[LMAX-Engine] ✅ Batch Flushed: %d transaksi disimpan persisten.\n", len(logsToInsert))
	}
}
