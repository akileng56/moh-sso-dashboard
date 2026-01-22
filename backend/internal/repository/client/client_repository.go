package client

import (
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/keycloak"
	models "github.com/moh-sso-dashboard/internal/model"
)

type ClientRepository interface {
	// ------------------------------------------------
	// Client lifecycle
	// ------------------------------------------------
	CreateClient(app *models.Client) (string, error)
	GetClientByID(id uuid.UUID) (*models.Client, error)
	GetClientByClientID(clientID string) (*models.Client, error)
	ListClients() ([]models.Client, error)
	UpdateClient(app *models.Client) error
	DeleteClient(id uuid.UUID) error

	// ------------------------------------------------
	// Client roles / permissions
	// ------------------------------------------------
	CreateClientRole(clientID uuid.UUID, payload *models.CreateClientRoleRequest) error
	ListClientRoles(clientID uuid.UUID) ([]keycloak.ClientRoleRep, error)
	DeleteClientRole(clientID uuid.UUID, role string) error

	// ------------------------------------------------
	// User ↔ Client role mapping
	// ------------------------------------------------
	AssignClientRoleToUser(
		userID uuid.UUID,
		clientID uuid.UUID,
		role string,
	) error

	RemoveClientRoleFromUser(
		userID uuid.UUID,
		clientID uuid.UUID,
		role string,
	) error
}
