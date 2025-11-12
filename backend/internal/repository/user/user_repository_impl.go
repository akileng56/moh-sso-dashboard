package repository

import (
	"database/sql"

	models "github.com/moh-sso-dashboard/internal/model"
)

// dbUserRepository is a concrete implementation of the UserRepository interface.
// It holds a reference to a database connection.
type dbUserRepository struct {
	DB *sql.DB
}

// NewUserRepository is the "constructor" function.
// It takes the database connection as a dependency...
func NewUserRepository(db *sql.DB) UserRepository {
	return &dbUserRepository{
		DB: db,
	}
}

// CreateUser implements the UserRepository interface.
func (r *dbUserRepository) CreateUser(user *models.User) error {

	return nil
}

// GetUserByID implements the UserRepository interface.
func (r *dbUserRepository) GetUserByID(id string) (*models.User, error) {

	return nil, nil
}

// ListUsers implements the UserRepository interface.
func (r *dbUserRepository) ListUsers() ([]models.User, error) {

	return nil, nil
}

// UpdateUser implements the UserRepository interface.
func (r *dbUserRepository) UpdateUser(user *models.User) error {

	return nil
}

// DeleteUser implements the UserRepository interface.
func (r *dbUserRepository) DeleteUser(id string) error {

	return nil
}
