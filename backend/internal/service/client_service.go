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

	if _, err := s.repo.CreateClient(newClient); err != nil {
		return nil, err
	}

	// 🔔 Notify admins
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
		Metadata: utils.MustJSON(map[string]interface{}{
			"client_id": client.ClientID,
			"name":      client.Name,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}

/*
|--------------------------------------------------------------------------
| Client roles / permissions
|--------------------------------------------------------------------------
*/

// CREATE CLIENT ROLE
func (s *ClientService) CreateClientRole(
	ctx context.Context,
	clientID uuid.UUID,
	role string,
	adminID uuid.UUID,
) error {

	if role == "" {
		return errors.New("role name is required")
	}

	if err := s.repo.CreateClientRole(clientID, role); err != nil {
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
			"role":      role,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}

// LIST CLIENT ROLES
func (s *ClientService) ListClientRoles(
	clientID uuid.UUID,
) ([]string, error) {
	return s.repo.ListClientRoles(clientID)
}

// DELETE CLIENT ROLE
func (s *ClientService) DeleteClientRole(
	ctx context.Context,
	clientID uuid.UUID,
	role string,
	adminID uuid.UUID,
) error {

	if err := s.repo.DeleteClientRole(clientID, role); err != nil {
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

/*
|--------------------------------------------------------------------------
| User ↔ Client role mapping
|--------------------------------------------------------------------------
*/

// ASSIGN ROLE TO USER
func (s *ClientService) AssignClientRoleToUser(
	ctx context.Context,
	userID uuid.UUID,
	clientID uuid.UUID,
	role string,
	adminID uuid.UUID,
) error {

	if err := s.repo.AssignClientRoleToUser(userID, clientID, role); err != nil {
		return err
	}

	nt := models.ClientRoleAssigned
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Client role assigned to user",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"user_id":   userID.String(),
			"client_id": clientID.String(),
			"role":      role,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}

// REMOVE ROLE FROM USER
func (s *ClientService) RemoveClientRoleFromUser(
	ctx context.Context,
	userID uuid.UUID,
	clientID uuid.UUID,
	role string,
	adminID uuid.UUID,
) error {

	if err := s.repo.RemoveClientRoleFromUser(userID, clientID, role); err != nil {
		return err
	}

	nt := models.ClientRoleRemoved
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Client role removed from user",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"user_id":   userID.String(),
			"client_id": clientID.String(),
			"role":      role,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}
