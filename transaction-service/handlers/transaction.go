package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"transaction-service/config"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type TransferRequest struct {
	FromAccountID uuid.UUID `json:"from_account_id"`
	ToAccountID   uuid.UUID `json:"to_account_id"`
	Amount        float64   `json:"amount"`
}

// Transfer kini bersifat ASINKRON:
// 1. Validasi request (cepat, di memori)
// 2. Masukkan ke antrean Kafka
// 3. Balas HTTP 202 Accepted langsung ke client
// => Eksekusi ACID yang berat diproses oleh transaction-worker di belakang layar
func Transfer(c *fiber.Ctx) error {
	var req TransferRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid Request Body Format"})
	}

	// Validasi input (tetap sinkron, sangat cepat, tidak menyentuh DB)
	if req.FromAccountID == uuid.Nil {
		return c.Status(400).JSON(fiber.Map{"error": "FROM_ACCOUNT parse failed (Nil)"})
	}
	if req.ToAccountID == uuid.Nil {
		return c.Status(400).JSON(fiber.Map{"error": "TO_ACCOUNT parse failed (Nil)"})
	}
	if req.Amount <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Amount must be strictly positive"})
	}
	if req.FromAccountID == req.ToAccountID {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot transfer to same self-account"})
	}

	// Buat pesan untuk dikirim ke Kafka
	referenceNumber := fmt.Sprintf("TRX-%d", time.Now().UnixNano())
	message := map[string]interface{}{
		"reference_number": referenceNumber,
		"from_account_id":  req.FromAccountID.String(),
		"to_account_id":    req.ToAccountID.String(),
		"amount":           req.Amount,
		"submitted_at":     time.Now().UTC().Format(time.RFC3339),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyiapkan pesan transaksi"})
	}

	// Kirim ke Kafka (non-blocking bagi client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = config.KafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(referenceNumber),
		Value: messageBytes,
	})
	if err != nil {
		return c.Status(503).JSON(fiber.Map{
			"status": "QUEUE_ERROR",
			"error":  "Sistem antrean sementara tidak tersedia, coba lagi",
		})
	}

	// Langsung balas 202 Accepted tanpa menunggu DB
	return c.Status(202).JSON(fiber.Map{
		"status":           "ACCEPTED",
		"reference_number": referenceNumber,
		"message":          "Transfer sedang diproses di antrean. Gunakan reference_number untuk mengecek status.",
	})
}
