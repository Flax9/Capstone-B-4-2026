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
