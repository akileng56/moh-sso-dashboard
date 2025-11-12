package repository

import (
	models "github.com/moh-sso-dashboard/internal/model"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	ListUsers() ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id string) error
}
