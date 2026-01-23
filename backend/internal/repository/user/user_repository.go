package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/keycloak"
	models "github.com/moh-sso-dashboard/internal/model"
)

type UserRepository interface {
	CreateUser(user *models.User) (string, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	ListUsers() ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id string) error
	ToggleUserEnabled(
		ctx context.Context,
		userID string,
		enabled bool,
	) error

	ResetUserPassword(
		ctx context.Context,
		userID string,
	) error

	GetUserClientRolesForClient(
		ctx context.Context,
		userID string,
		clientID string,
		clientUUID string,
	) ([]keycloak.ClientRoleRep, error)

	GetUserClientRoles(
		ctx context.Context,
		userID string,
	) ([]keycloak.UserClientRoleAssignment, error)

	AddUserClientRoles(
		ctx context.Context,
		userID string,
		clientID string,
		clientUUID string,
		roles []string,
	) error

	RemoveUserClientRoles(
		ctx context.Context,
		userID string,
		clientID string,
		clientUUID string,
		roles []string,
	) error
}
