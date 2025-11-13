package service

import (
	"errors"
	"fmt"

	models "github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/user"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserRequest is the DTO for creating a new user.
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

// UserService handles business logic for users.
type UserService struct {
	repo repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// CreateUser validates, hashes the password, and creates a new user.
func (s *UserService) CreateUser(req CreateUserRequest) (*models.User, error) {
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return nil, errors.New("username, email, and password are required")
	}
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}
	_, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Transform DTO (CreateUserRequest) into a Model (User)
	newUser := &models.User{
		Username: req.Username,
		Email:    req.Email,
		// TODO: Set other defaults (e.g., IsActive: true)
	}

	// Call the repository to save the new user
	err = s.repo.CreateUser(newUser)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

// GetUser retrieves a user by their ID.
func (s *UserService) GetUser(id string) (*models.User, error) {
	return s.repo.GetUserByID(id)
}

// ListUsers retrieves all users.
func (s *UserService) ListUsers() ([]models.User, error) {
	return s.repo.ListUsers()
}

// DeleteUser deletes a user by their ID.
func (s *UserService) DeleteUser(id string) error {
	return s.repo.DeleteUser(id)
}
