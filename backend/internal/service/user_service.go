package service

import (
	"errors"

	"github.com/google/uuid"
	models "github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/user"
)

type CreateUserRequest struct {
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	FirstName        string              `json:"first_name,omitempty"`
	LastName         string              `json:"last_name,omitempty"`
	FullName         string              `json:"full_name,omitempty"`          // convenience input
	Enabled          *bool               `json:"enabled,omitempty"`            // default true
	RequirePwdChange *bool               `json:"require_pwd_change,omitempty"` // force reset on first login
	Password         string              `json:"password,omitempty"`           // optional if using invite flow
	SendInvite       bool                `json:"send_invite,omitempty"`
	RealmRoles       []string            `json:"realm_roles,omitempty"`  // e.g. ["admin"]
	ClientRoles      map[string][]string `json:"client_roles,omitempty"` // client_id -> roles[]
}

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(req CreateUserRequest) (*models.User, error) {

	if req.Username == "" || req.Email == "" {
		return nil, errors.New("username and email are required")
	}

	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Enabled:   true,
	}

	if _, err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
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
