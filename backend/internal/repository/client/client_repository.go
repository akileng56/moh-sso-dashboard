package repository

import (
	models "github.com/moh-sso-dashboard/internal/model"
)

type ClientRepository interface {
	CreateClient(app *models.AppRegistry) error
	GetClientByID(id string) (*models.AppRegistry, error)
	ListClients() ([]models.AppRegistry, error)
	UpdateClient(app *models.AppRegistry) error
	DeleteClient(id string) error
}
