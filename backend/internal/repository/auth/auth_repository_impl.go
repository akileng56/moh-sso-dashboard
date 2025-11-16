package auth

import (
	"errors"
	"fmt"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/keycloak"
)

// keycloakAuthRepository is the implementation that uses the keycloak.Client
type keycloakAuthRepository struct {
	keycloakClient *keycloak.Client
	redirectURI    string
}

// NewAuthRepository creates a new Keycloak-backed authentication repository
func NewAuthRepository(kcClient *keycloak.Client, config config.Config) AuthRepository {
	return &keycloakAuthRepository{
		keycloakClient: kcClient,
		redirectURI:    config.KeycloakRedirectUri,
	}
}

func (r *keycloakAuthRepository) ExchangeCode(code string) (*keycloak.TokenResponse, error) {
	if r.redirectURI == "" {
		return nil, errors.New("KEYCLOAK_REDIRECT_URI environment variable is not set")
	}

	// Use the client method, passing the required code and the public redirect URI
	tokens, err := r.keycloakClient.ExchangeCodeForToken(code, r.redirectURI)
	if err != nil {
		return nil, fmt.Errorf("keycloak exchange failed: %w", err)
	}

	return tokens, nil
}
