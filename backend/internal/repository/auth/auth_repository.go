package auth

import "github.com/moh-sso-dashboard/internal/keycloak"

type AuthRepository interface {
	ExchangeCode(code string) (*keycloak.TokenResponse, error)
}
