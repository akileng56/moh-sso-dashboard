package client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
	models "github.com/moh-sso-dashboard/internal/model"
)

type sqlcClientRepository struct {
	keycloakClient *keycloak.Client
	config         *config.Config
	db             db.Store
	logger         *logger.Logger
}

func NewClientRepository(
	keycloakClient *keycloak.Client,
	config *config.Config,
	db db.Store,
	log logger.Logger,
) ClientRepository {

	// Authenticate Keycloak admin client
	if err := keycloakClient.Authenticate(); err != nil {
		fmt.Errorf("Failed to authenticate Keycloak admin client: %v", err)
		panic(fmt.Sprintf("Keycloak authentication failed: %v", err))
	}

	return &sqlcClientRepository{
		keycloakClient: keycloakClient,
		config:         config,
		db:             db,
		logger:         &log,
	}
}

func (r *sqlcClientRepository) CreateClient(client *models.Client) error {
	ctx := context.Background()

	existingKCClient, _ := r.keycloakClient.GetClientByClientID(client.ClientID)
	if existingKCClient != nil {
		return errors.New("client already exists in keycloak")
	}

	err := r.keycloakClient.CreateClient(keycloak.CreateClientParams{
		ClientID:     client.ClientID,
		Name:         client.Name,
		Description:  client.Description,
		BaseURL:      client.BaseURL,
		PublicClient: client.PublicClient,
	})
	if err != nil {
		return fmt.Errorf("failed to create in keycloak: %w", err)
	}

	return r.db.CreateClient(ctx, db.CreateClientParams{
		ClientID: client.ClientID,
		Name:     client.Name,
		Description: sql.NullString{
			String: client.Description,
			Valid:  client.Description != "",
		},
		BaseUrl: sql.NullString{
			String: client.BaseURL,
			Valid:  client.BaseURL != "",
		},
		Icon: sql.NullString{
			String: client.Attributes["icon"],
			Valid:  client.Attributes["icon"] != "",
		},
		PublicClient: sql.NullBool{
			Bool:  client.PublicClient,
			Valid: client.PublicClient,
		},
		Enabled: sql.NullBool{
			Bool:  client.Enabled,
			Valid: client.Enabled,
		},
	})
}

// GetClientByID retrieves a client by its ID.
func (r *sqlcClientRepository) GetClientByID(id string) (*models.Client, error) {
	ctx := context.Background()

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	row, err := r.db.GetClientByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &models.Client{
		ID:           row.ID.String(),
		ClientID:     row.ClientID,
		Name:         row.Name,
		Description:  row.Description.String,
		BaseURL:      row.BaseUrl.String,
		PublicClient: row.PublicClient.Bool,
		Enabled:      row.Enabled.Bool,
	}, nil
}

// ListClients retrieves all client entries.
func (r *sqlcClientRepository) ListClients() ([]models.Client, error) {
	ctx := context.Background()

	kcClients, err := r.keycloakClient.ListClients()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch keycloak clients: %w", err)
	}

	dbClients, err := r.db.ListClients(ctx)
	if err != nil {
		return nil, err
	}

	// Index DB by ClientID for easy merge
	dbMap := map[string]db.Client{}
	for _, c := range dbClients {
		dbMap[c.ClientID] = c
	}

	// 3. Merge KC + DB (KC is primary)
	var result []models.Client
	for _, kc := range kcClients {
		// Prepare the attributes map with just the "icon" key
		attributes := map[string]string{}
		if icon, ok := kc.Attributes["icon"]; ok {
			attributes["icon"] = icon
		}

		// If exists in DB, merge metadata
		if entry, ok := dbMap[kc.ClientID]; ok {
			result = append(result, models.Client{
				ID:           entry.ID.String(),
				ClientID:     kc.ClientID,
				Name:         kc.Name,
				Description:  kc.Description,
				BaseURL:      kc.BaseURL,
				PublicClient: kc.PublicClient,
				Enabled:      kc.Enabled,
				Attributes:   attributes,
			})
		} else {
			// KC-only record
			result = append(result, models.Client{
				ID:           "",
				ClientID:     kc.ClientID,
				Name:         kc.Name,
				Description:  kc.Description,
				BaseURL:      kc.BaseURL,
				PublicClient: kc.PublicClient,
				Enabled:      kc.Enabled,
				Attributes:   attributes,
			})
		}
	}

	return result, nil
}

func (r *sqlcClientRepository) UpdateClient(client *models.Client) error {
	return nil
}

func (r *sqlcClientRepository) DeleteClient(id uuid.UUID) error {
	ctx := context.Background()

	dbClient, err := r.db.GetClientByID(ctx, id)
	if err != nil {
		return fmt.Errorf("client not found: %w", err)
	}

	if err := r.keycloakClient.DeleteClient(dbClient.ClientID); err != nil {
		return fmt.Errorf("failed deleting keycloak client: %w", err)
	}

	return r.db.DeleteClient(ctx, id)
}
