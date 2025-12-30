package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	models "github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/user"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(req CreateUserRequest) (*models.User, error) {
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return nil, errors.New("username, email, and password are required")
	}
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	// Hash password (but then throw it away)
	_, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &models.User{
		ID:        "",
		Username:  req.Username,
		FirstName: req.FullName,
		LastName:  "",
		Email:     req.Email,
		Enabled:   true,
		Roles:     []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = s.repo.CreateUser(newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *UserService) GetUser(id uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *UserService) ListUsers() ([]models.User, error) {
	return s.repo.ListUsers()
}

func (s *UserService) DeleteUser(id string) error {
	return s.repo.DeleteUser(id)
}
