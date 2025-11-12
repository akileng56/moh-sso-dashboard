package service

import (
	models "github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/client"
)

// CreateClientRequest is the DTO for creating a new client.
// It contains only the fields a user is allowed to provide.
type CreateClientRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	RedirectURIs []string `json:"redirect_uris"`
	// Add other fields from AppRegistry that a user can set
}

// ClientService handles business logic for clients.
type ClientService struct {
	repo repository.ClientRepository
}

// NewClientService creates a new ClientService.
func NewClientService(repo repository.ClientRepository) *ClientService {
	return &ClientService{repo: repo}
}

// CreateClient validates the request, transforms it into a model,
// and asks the repository to save it.
func (s *ClientService) CreateClient(req CreateClientRequest) (*models.AppRegistry, error) {
	// Transform DTO (CreateClientRequest) into a Model (AppRegistry)
	newApp := &models.AppRegistry{
		Name:        req.Name,
		Description: req.Description,
		// TODO: Generate a ClientID, ClientSecret, etc.
		// Example:
		// ClientID:     generateMyClientID(),
		// ClientSecret: generateMyClientSecret(),
	}

	err := s.repo.CreateClient(newApp)
	if err != nil {
		return nil, err
	}

	return newApp, nil
}

// GetClient retrieves a client by its ID.
func (s *ClientService) GetClient(id string) (*models.AppRegistry, error) {
	return s.repo.GetClientByID(id)
}

// ListClients retrieves all clients.
func (s *ClientService) ListClients() ([]models.AppRegistry, error) {
	return s.repo.ListClients()
}

// DeleteClient deletes a client by its ID.
func (s *ClientService) DeleteClient(id string) error {
	return s.repo.DeleteClient(id)
}
