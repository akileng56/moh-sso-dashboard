package service

import (
	"fmt"

	"github.com/moh-sso-dashboard/internal/keycloak"
	repository "github.com/moh-sso-dashboard/internal/repository/auth"
)

// AuthService defines the business operations for authentication
type AuthService interface {
	ProcessAuthCode(code string) (*keycloak.TokenResponse, error)
}

// authService implementation
type authService struct {
	authRepo repository.AuthRepository
	// userRepo repository.UserRepository // Uncomment if you need to check/save users locally
}

// NewAuthService creates a new authentication service
func NewAuthService(authRepo repository.AuthRepository) AuthService {
	return &authService{
		authRepo: authRepo,
	}
}

// ProcessAuthCode handles the complete Authorization Code exchange process
func (s *authService) ProcessAuthCode(code string) (*keycloak.TokenResponse, error) {
	tokens, err := s.authRepo.ExchangeCode(code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// 2. 🚨 Business Logic (Example)
	// You can parse the ID Token here to get user info (email, roles)
	// and check if they exist in your local database.
	// userClaims := parseIDToken(tokens.IDToken)
	// user, err := s.userRepo.FindOrCreate(userClaims)

	// 3. Return the tokens to the handler for session management
	return tokens, nil
}
