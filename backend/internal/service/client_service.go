package service

import (
	"context"
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
	ClientSecret string   `json:"client_secret,omitempty"`
	RedirectURIs []string `json:"redirect_uris" validate:"required,dive,uri"`
	WebOrigins   []string `json:"web_origins,omitempty"`

	// Flow Controls
	StandardFlowEnabled    bool `json:"standard_flow_enabled"`
	ImplicitFlowEnabled    bool `json:"implicit_flow_enabled"`
	DirectAccessGrants     bool `json:"direct_access_grants"`
	ServiceAccountsEnabled bool `json:"service_accounts_enabled"`

	// Access & Security
	PublicClient bool `json:"public_client"`

	// Optional Settings
	RootURL   string `json:"root_url,omitempty"`
	BaseURL   string `json:"base_url,omitempty"`
	AdminURL  string `json:"admin_url,omitempty"`
	Enabled   bool   `json:"enabled"`
	Protocol  string `json:"protocol,omitempty"`
	LoginURI  string `json:"login_uri,omitempty"`
	LogoutURI string `json:"logout_uri,omitempty"`

	// Roles
	DefaultClientScopes  []string `json:"default_client_scopes,omitempty"`
	OptionalClientScopes []string `json:"optional_client_scopes,omitempty"`

	// Metadata
	Tags []string `json:"tags,omitempty"`
}

type ClientService struct {
	repo          repository.ClientRepository
	notifications NotificationsService
}

func NewClientService(
	repo repository.ClientRepository,
	notifications NotificationsService,
) *ClientService {
	return &ClientService{
		repo:          repo,
		notifications: notifications,
	}
}

// ----------------------------------------------------
// CREATE CLIENT
// ----------------------------------------------------
func (s *ClientService) CreateClient(
	ctx context.Context,
	req CreateClientRequest,
	adminID uuid.UUID,
) (*models.Client, error) {

	if req.Name == "" {
		return nil, errors.New("client name is required")
	}

	clientID := req.ClientID
	if clientID == "" {
		clientID = utils.GenerateClientID()
	}

	newClient := &models.Client{
		ClientID:     clientID,
		Name:         req.Name,
		Description:  req.Description,
		BaseURL:      req.BaseURL,
		PublicClient: req.PublicClient,
		Enabled:      true,
	}

	if err := s.repo.CreateClient(newClient); err != nil {
		return nil, err
	}

	// 🔔 Notification: Client Created (info)
	nt := models.ClientCreated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "New client application created",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"client_id": newClient.ClientID,
			"name":      newClient.Name,
			"admin_id":  adminID.String(),
		}),
	})

	return newClient, nil
}

// ----------------------------------------------------
// GET CLIENT
// ----------------------------------------------------
func (s *ClientService) GetClient(id string) (*models.Client, error) {
	return s.repo.GetClientByID(id)
}

// ----------------------------------------------------
// LIST CLIENTS
// ----------------------------------------------------
func (s *ClientService) ListClients() ([]models.Client, error) {
	return s.repo.ListClients()
}

// ----------------------------------------------------
// DELETE CLIENT (CRITICAL)
// ----------------------------------------------------
func (s *ClientService) DeleteClient(
	ctx context.Context,
	id uuid.UUID,
	adminID uuid.UUID,
) error {

	client, err := s.repo.GetClientByID(id.String())
	if err != nil {
		return err
	}

	if err := s.repo.DeleteClient(id); err != nil {
		return err
	}

	// 🔔 Notification: Client Deleted (critical)
	nt := models.ClientDeleted
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Client application deleted",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"client_id": client.ClientID,
			"name":      client.Name,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}
