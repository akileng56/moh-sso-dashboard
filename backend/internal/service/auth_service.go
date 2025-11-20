package service

import (
	"fmt"

	"github.com/moh-sso-dashboard/internal/keycloak"
	repository "github.com/moh-sso-dashboard/internal/repository/auth"
)

type AuthService interface {
	ProcessAuthCode(code string) (*keycloak.TokenResponse, error)
	GetAccessToken(refreshToken string) (*keycloak.TokenResponse, error)
	GetMe(accessToken string) (*keycloak.AuthUser, error)
	LogOut(efreshToken string) error
}

// authService implementation
type authService struct {
	authRepo repository.AuthRepository
}

func NewAuthService(authRepo repository.AuthRepository) AuthService {
	return &authService{
		authRepo: authRepo,
	}
}

func (s *authService) ProcessAuthCode(code string) (*keycloak.TokenResponse, error) {
	tokens, err := s.authRepo.ExchangeCode(code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return tokens, nil
}

func (s *authService) GetAccessToken(refreshToken string) (*keycloak.TokenResponse, error) {

	accessToken, err := s.authRepo.GetAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("Failed to get access token : %v", err)
	}
	return accessToken, nil
}

func (s *authService) GetMe(accessToken string) (*keycloak.AuthUser, error) {

	userProfile, err := s.authRepo.GetMe(accessToken)

	if err != nil {
		return nil, fmt.Errorf("failed to get user profile")
	}
	return userProfile, nil
}

func (s *authService) LogOut(refreshToken string) error {
	err := s.authRepo.Logout(refreshToken)
	if err != nil {
		return fmt.Errorf("failed to logout user ")
	}
	return nil

}
