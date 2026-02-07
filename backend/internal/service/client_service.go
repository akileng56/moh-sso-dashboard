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
	ClientID               string            `json:"clientId" validate:"required"`
	ClientSecret           string            `json:"clientSecret"`
	RedirectURIs           []string          `json:"redirectUris" validate:"required,dive,uri"`
	WebOrigins             []string          `json:"webOrigins"`
	StandardFlowEnabled    bool              `json:"standardFlowEnabled"`
	ImplicitFlowEnabled    bool              `json:"implicitFlowEnabled"`
	DirectAccessGrants     bool              `json:"directAccessGrants"`
	ServiceAccountsEnabled bool              `json:"serviceAccountsEnabled"`
	PublicClient           bool              `json:"publicClient"`
	RootURL                string            `json:"rootUrl"`
	BaseURL                string            `json:"baseUrl"`
	AdminURL               string            `json:"adminUrl"`
	Enabled                bool              `json:"enabled"`
	Protocol               string            `json:"protocol,omitempty"`
	LoginURI               string            `json:"loginUri,omitempty"`
	LogoutURI              string            `json:"logoutUri,omitempty"`
	Attributes             map[string]string `json:"attributes"`
	DefaultClientScopes    []string          `json:"defaultClientScopes,omitempty"`
	OptionalClientScopes   []string          `json:"optionalClientScopes,omitempty"`
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
		RootURL:      req.RootURL,
		RedirectUris: req.RedirectURIs,
		WebOrigins:   req.WebOrigins,
		PublicClient: req.PublicClient,
		Enabled:      req.Enabled,
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
