package model

import (
	"time"
)

type Client struct {
	ID           string `json:"id"`
	ClientID     string `json:"clientId"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	BaseURL      string `json:"baseUrl"`
	Icon         string `json:"icon"`
	PublicClient bool   `json:"publicClient"`
	Enabled      bool   `json:"enabled"`
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
