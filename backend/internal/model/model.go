package model

import (
	"time"
)

type AppRegistry struct {
	ID          string    `gorm:"primaryKey"`
	ClientID    string    // Keycloak client ID
	Name        string
	Description string
	URL         string
	OwnerID     string    // Keycloak user ID of app owner
	Status      string    // e.g. "active", "pending", "disabled"
	CreatedAt   time.Time
	UpdatedAt   time.Time
}


type User struct {
	ID        string    `gorm:"primaryKey"`
	Username  string
	Email     string
	KeycloakID string   // maps to the real Keycloak user ID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserResponse is a "safe" version of the User model for JSON responses.
// Notice it omits the HashedPassword.
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	IsActive bool   `json:"is_active"`
}
