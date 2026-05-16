package models

import (
	"time"
	"github.com/google/uuid"
)

type User struct {
	UserID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"user_id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	FullName     string    `gorm:"not null" json:"full_name"`
	TokenVersion int       `gorm:"default:1" json:"token_version"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
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
