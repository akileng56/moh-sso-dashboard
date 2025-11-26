package service

import (
	"errors"

	"github.com/google/uuid"
	models "github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/client"
	"github.com/moh-sso-dashboard/internal/utils"
)

type CreateClientRequest struct {
	// Basic Info
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`

	// OAuth / OIDC Information
	ClientID     string   `json:"client_id" validate:"required"`
	ClientSecret string   `json:"client_secret,omitempty"` // only for confidential clients
	RedirectURIs []string `json:"redirect_uris" validate:"required,dive,uri"`
	WebOrigins   []string `json:"web_origins,omitempty"`

	// Flow Controls
	StandardFlowEnabled    bool `json:"standard_flow_enabled"`    // Authorization Code
	ImplicitFlowEnabled    bool `json:"implicit_flow_enabled"`    // Implicit
	DirectAccessGrants     bool `json:"direct_access_grants"`     // Resource Owner Password
	ServiceAccountsEnabled bool `json:"service_accounts_enabled"` // For backend-to-backend

	// Access & Security
	PublicClient bool `json:"public_client"` // true = no secret required

	// Optional Settings
	RootURL   string `json:"root_url,omitempty"`
	BaseURL   string `json:"base_url,omitempty"`
	AdminURL  string `json:"admin_url,omitempty"`
	Enabled   bool   `json:"enabled"`            // enable/disable client
	Protocol  string `json:"protocol,omitempty"` // default "openid-connect"`
	LoginURI  string `json:"login_uri,omitempty"`
	LogoutURI string `json:"logout_uri,omitempty"`

	// Roles
	DefaultClientScopes  []string `json:"default_client_scopes,omitempty"`
	OptionalClientScopes []string `json:"optional_client_scopes,omitempty"`

	// Metadata
	Tags []string `json:"tags,omitempty"`
}

type ClientService struct {
	repo repository.ClientRepository
}

func NewClientService(repo repository.ClientRepository) *ClientService {
	return &ClientService{repo: repo}
}

func (s *ClientService) CreateClient(req CreateClientRequest) (*models.Client, error) {
	if req.Name == "" {
		return nil, errors.New("client name is required")
	}

	clientID := req.ClientID
	if clientID == "" {
		clientID = utils.GenerateClientID()
	}

	newClient := &models.Client{
		ID:           uuid.New().String(),
		ClientID:     clientID,
		Name:         req.Name,
		Description:  req.Description,
		BaseURL:      req.BaseURL,
		PublicClient: req.PublicClient,
		Enabled:      true, // always true on creation
	}

	if err := s.repo.CreateClient(newClient); err != nil {
		return nil, err
	}

	return newClient, nil
}

func (s *ClientService) GetClient(id string) (*models.Client, error) {
	return s.repo.GetClientByID(id)
}

func (s *ClientService) ListClients() ([]models.Client, error) {
	return s.repo.ListClients()
}

func (s *ClientService) DeleteClient(id uuid.UUID) error {
	return s.repo.DeleteClient(id)
}
