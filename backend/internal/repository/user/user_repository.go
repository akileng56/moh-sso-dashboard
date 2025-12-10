package repository

import (
	"github.com/google/uuid"
	models "github.com/moh-sso-dashboard/internal/model"
)

type UserRepository interface {
	CreateUser(user *models.User) (string, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	ListUsers() ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id string) error
}
