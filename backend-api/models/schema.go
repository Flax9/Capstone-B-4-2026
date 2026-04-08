package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"user_id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"` // - (ignore in json output)
	FullName     string    `gorm:"not null" json:"full_name"`
	TokenVersion int       `gorm:"default:1" json:"token_version"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

type Account struct {
	AccountID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"account_id"`
	UserID        uuid.UUID `gorm:"type:uuid" json:"user_id"`
	AccountNumber string    `gorm:"uniqueIndex;not null" json:"account_number"`
	Balance       float64   `gorm:"type:numeric(15,2);default:0.00" json:"balance"`
	Currency      string    `gorm:"default:'IDR'" json:"currency"`
	UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
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

type AuditLog struct {
	LogID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"log_id"`
	UserID    uuid.UUID `gorm:"type:uuid" json:"user_id"`
	Action    string    `gorm:"not null" json:"action"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Details   string    `gorm:"type:jsonb" json:"details"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}
