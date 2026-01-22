package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/keycloak"
	models "github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/client"
	"github.com/moh-sso-dashboard/internal/utils"
)

type CreateClientRequest struct {
	Name                   string            `json:"name" validate:"required"`
	Description            string            `json:"description,omitempty"`
	ClientID               string            `json:"client_id" validate:"required"`
	ClientSecret           string            `json:"client_secret,omitempty"`
	RedirectURIs           []string          `json:"redirect_uris" validate:"required,dive,uri"`
	WebOrigins             []string          `json:"web_origins,omitempty"`
	StandardFlowEnabled    bool              `json:"standard_flow_enabled"`
	ImplicitFlowEnabled    bool              `json:"implicit_flow_enabled"`
	DirectAccessGrants     bool              `json:"direct_access_grants"`
	ServiceAccountsEnabled bool              `json:"service_accounts_enabled"`
	PublicClient           bool              `json:"public_client"`
	RootURL                string            `json:"root_url"`
	BaseURL                string            `json:"base_url"`
	AdminURL               string            `json:"admin_url"`
	Enabled                bool              `json:"enabled"`
	Protocol               string            `json:"protocol,omitempty"`
	LoginURI               string            `json:"login_uri,omitempty"`
	LogoutURI              string            `json:"logout_uri,omitempty"`
	Attributes             map[string]string `json:"attributes"`
	DefaultClientScopes    []string          `json:"default_client_scopes,omitempty"`
	OptionalClientScopes   []string          `json:"optional_client_scopes,omitempty"`
	Tags                   []string          `json:"tags,omitempty"`
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

func (s *ClientService) CreateClient(
	ctx context.Context,
	req CreateClientRequest,
	adminID uuid.UUID,
) (*models.Client, error) {

	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("client name is required")
	}

	clientID := req.ClientID
	if clientID == "" {
		clientID = utils.GenerateClientID()
	}

	client := &models.Client{
		ClientID:     clientID,
		Name:         req.Name,
		Description:  req.Description,
		BaseURL:      req.BaseURL,
		PublicClient: req.PublicClient,
		Enabled:      true,
		Attributes:   req.Attributes,
	}

	if _, err := s.repo.CreateClient(client); err != nil {
		return nil, err
	}

	nt := models.ClientCreated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "New client application created",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"client_id": client.ClientID,
			"name":      client.Name,
			"admin_id":  adminID.String(),
		}),
	})

	return client, nil
}

func (s *ClientService) GetClient(id uuid.UUID) (*models.Client, error) {
	return s.repo.GetClientByID(id)
}

func (s *ClientService) ListClients() ([]models.Client, error) {
	return s.repo.ListClients()
}

func (s *ClientService) DeleteClient(
	ctx context.Context,
	id uuid.UUID,
	adminID uuid.UUID,
) error {

	client, err := s.repo.GetClientByID(id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteClient(id); err != nil {
		return err
	}

	nt := models.ClientDeleted
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Client application deleted",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"client_id": client.ClientID,
			"name":      client.Name,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}

func (s *ClientService) ToggleClientEnabled(
	ctx context.Context,
	clientID uuid.UUID,
	enabled bool,
	adminID uuid.UUID,
) error {

	if err := s.repo.ToggleClientEnabled(ctx, clientID, enabled); err != nil {
		return err
	}

	var nt models.NotificationType
	msg := "Client updated"

	if enabled {
		nt = models.ClientEnabled
		msg = "Client enabled"
	} else {
		nt = models.ClientDisabled
		msg = "Client disabled"
	}

	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    msg,
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"client_id": clientID.String(),
			"enabled":   enabled,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}

// CREATE ROLE
func (s *ClientService) CreateClientRole(
	ctx context.Context,
	clientID uuid.UUID,
	payload *models.CreateClientRoleRequest,
	adminID uuid.UUID,
) error {

	if strings.TrimSpace(payload.Role) == "" {
		return errors.New("role name is required")
	}

	if err := s.repo.CreateClientRole(ctx, clientID, payload); err != nil {
		return err
	}

	nt := models.ClientRoleCreated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Client role created",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"client_id": clientID.String(),
			"role":      payload.Role,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}

// LIST ROLES
func (s *ClientService) ListClientRoles(
	ctx context.Context,
	clientID uuid.UUID,
) ([]keycloak.ClientRoleRep, error) {

	return s.repo.ListClientRoles(ctx, clientID)
}

// DELETE ROLE
func (s *ClientService) DeleteClientRole(
	ctx context.Context,
	clientID uuid.UUID,
	role string,
	adminID uuid.UUID,
) error {

	if err := s.repo.DeleteClientRole(ctx, clientID, role); err != nil {
		return err
	}

	nt := models.ClientRoleDeleted
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Client role deleted",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"client_id": clientID.String(),
			"role":      role,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}
