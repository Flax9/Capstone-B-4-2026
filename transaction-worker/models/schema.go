package models

import (
	"time"
	"github.com/google/uuid"
)

type Account struct {
	AccountID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"account_id"`
	UserID        uuid.UUID `gorm:"type:uuid" json:"user_id"`
	AccountNumber string    `gorm:"uniqueIndex;not null" json:"account_number"`
	Balance       float64   `gorm:"type:numeric(15,2);default:0.00" json:"balance"`
	Currency      string    `gorm:"default:'IDR'" json:"currency"`
	UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type Transaction struct {
	TransactionID   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"transaction_id"`
	ReferenceNumber string    `gorm:"uniqueIndex;not null" json:"reference_number"`
	FromAccountID   uuid.UUID `gorm:"type:uuid" json:"from_account_id"`
	ToAccountID     uuid.UUID `gorm:"type:uuid" json:"to_account_id"`
	Amount          float64   `gorm:"type:numeric(15,2);not null" json:"amount"`
	TransactionType string    `gorm:"not null" json:"transaction_type"`
	Status          string    `gorm:"default:'PENDING'" json:"status"`
	CreatedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}
