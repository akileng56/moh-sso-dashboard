package model

import (
	"time"
)

type AppRegistry struct {
	ID          string `gorm:"primaryKey"`
	Name        string
	Description string
	URL         string
	OwnerID     string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type User struct {
	ID         string `gorm:"primaryKey"`
	Username   string
	Email      string
	KeycloakID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	IsActive bool   `json:"is_active"`
}
