package client

import (
	"github.com/google/uuid"
	models "github.com/moh-sso-dashboard/internal/model"
)

type ClientRepository interface {
	CreateClient(app *models.Client) error
	GetClientByID(id string) (*models.Client, error)
	ListClients() ([]models.Client, error)
	UpdateClient(app *models.Client) error
	DeleteClient(id uuid.UUID) error
}
