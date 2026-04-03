package main

import (
	"time"

	"gorm.io/gorm"
)

// User is the database model for the user-service.
type User struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name         string    `gorm:"not null"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"` // soft-delete
}
