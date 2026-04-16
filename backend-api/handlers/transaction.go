package handlers

import (
	"banking-backend/config"
	"banking-backend/models"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransferRequest struct {
	FromAccountID uuid.UUID `json:"from_account_id"`
	ToAccountID   uuid.UUID `json:"to_account_id"`
	Amount        float64   `json:"amount"`
}

func Transfer(c *fiber.Ctx) error {
	var req TransferRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid Request Body Format"})
	}

	// ---> SUNTIKAN DEBUG: Cek kegagalan parse JSON UUID
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

	// Memulai Transaksional Keras (Atomicity penuh)
	// => Terjamin 100% dipaksa me-rute ke [MASTER DB]
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var fromAccount models.Account
		var toAccount models.Account

		// (Pessimistic Locking 'FOR UPDATE') - Kebal Race Condition!
		// Menggunakan AST asli GORM agar PostgreSQL tidak memuntahkan Syntax Error terkait LIMIT 1
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("account_id = ?", req.FromAccountID).First(&fromAccount).Error; err != nil {
			return fmt.Errorf("sender_account_not_found: %v", err)
		}

		if fromAccount.Balance < req.Amount {
			return fmt.Errorf("insufficient_balance")
		}

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("account_id = ?", req.ToAccountID).First(&toAccount).Error; err != nil {
			return fmt.Errorf("receiver_account_not_found: %v", err)
		}

		// Manipulasi Saldo (Memory RAM Lokal)
		fromAccount.Balance -= req.Amount
		toAccount.Balance += req.Amount

		// Dorong Pembaruan (Disk I/O)
		if err := tx.Save(&fromAccount).Error; err != nil { return err }
		if err := tx.Save(&toAccount).Error; err != nil { return err }

		// Rekam Log Transaksi
		transactionLog := models.Transaction{
			ReferenceNumber: fmt.Sprintf("TRX-%d", time.Now().UnixMilli()),
			FromAccountID:   req.FromAccountID,
			ToAccountID:     req.ToAccountID,
			Amount:          req.Amount,
			TransactionType: "TRANSFER",
			Status:          "SUCCESS",
		}
		if err := tx.Create(&transactionLog).Error; err != nil { return err }

		return nil
	})

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "FAILED", "error": err.Error()})
	}

	// Operasi Sukses! Eksekusi INVALIDASI REDIS
	// Menghancurkan record statis agar penarikan balance berikutnya fresh dari DB
	config.RedisClient.Del(config.Ctx, fmt.Sprintf("account:balance:%s", req.FromAccountID.String()))
	config.RedisClient.Del(config.Ctx, fmt.Sprintf("account:balance:%s", req.ToAccountID.String()))

	return c.Status(200).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Transfer diotentikasi & diverifikasi tuntas",
	})
}