package client

import (
	"context"
	"database/sql"
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

	if err := keycloakClient.Authenticate(); err != nil {
		panic(fmt.Sprintf("Keycloak authentication failed: %v", err))
	}

	return &sqlcClientRepository{
		keycloakClient: keycloakClient,
		config:         config,
		db:             db,
		logger:         &log,
	}
}

// CREATE (Keycloak = Source of Truth)
func (r *sqlcClientRepository) CreateClient(client *models.Client) error {
	ctx := context.Background()

	// 1️⃣ Create in Keycloak
	if err := r.keycloakClient.CreateClient(keycloak.CreateClientParams{
		ClientID:     client.ClientID,
		Name:         client.Name,
		Description:  client.Description,
		BaseURL:      client.BaseURL,
		PublicClient: client.PublicClient,
		Protocol:     "openid-connect",
		RedirectURIs: []string{},
		WebOrigins:   []string{},
	}); err != nil {
		return fmt.Errorf("keycloak create failed: %w", err)
	}

	// 2️⃣ Read back authoritative KC state
	kcClient, err := r.keycloakClient.GetClientByClientID(client.ClientID)
	if err != nil {
		return fmt.Errorf("keycloak read-back failed: %w", err)
	}

	// 3️⃣ Best-effort DB UPSERT
	if err := r.db.UpsertClient(ctx, db.UpsertClientParams{
		ClientID: kcClient.ClientID,
		Name:     kcClient.Name,
		Description: sql.NullString{
			String: kcClient.Description,
			Valid:  kcClient.Description != "",
		},
		BaseUrl: sql.NullString{
			String: kcClient.BaseURL,
			Valid:  kcClient.BaseURL != "",
		},
		PublicClient: sql.NullBool{
			Bool:  kcClient.PublicClient,
			Valid: true,
		},
		Enabled: sql.NullBool{
			Bool:  kcClient.Enabled,
			Valid: true,
		},
	}); err != nil {
		r.logger.Warn(
			"client created in keycloak but failed to sync db",
			"clientId", kcClient.ClientID,
			"error", err,
		)
	}

	return nil
}

// GET BY ID (Keycloak authoritative)
func (r *sqlcClientRepository) GetClientByID(id string) (*models.Client, error) {
	ctx := context.Background()

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	dbRow, err := r.db.GetClientByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	kcClient, err := r.keycloakClient.GetClientByClientID(dbRow.ClientID)
	if err != nil {
		return nil, fmt.Errorf("keycloak client not found: %w", err)
	}

	return &models.Client{
		ID:           dbRow.ID.String(),
		ClientID:     kcClient.ClientID,
		Name:         kcClient.Name,
		Description:  kcClient.Description,
		BaseURL:      kcClient.BaseURL,
		PublicClient: kcClient.PublicClient,
		Enabled:      kcClient.Enabled,
	}, nil
}

// LIST (KC primary, DB merged)
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

	dbMap := map[string]db.Client{}
	for _, c := range dbClients {
		dbMap[c.ClientID] = c
	}

	var result []models.Client

	for _, kc := range kcClients {
		attributes := map[string]string{}
		if icon, ok := kc.Attributes["icon"]; ok {
			attributes["icon"] = icon
		}

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

// UPDATE (KC first, DB best-effort)
func (r *sqlcClientRepository) UpdateClient(client *models.Client) error {
	ctx := context.Background()

	if err := r.keycloakClient.UpdateClient(client.ID, &models.Client{
		ClientID:     client.ClientID,
		Name:         client.Name,
		Description:  client.Description,
		BaseURL:      client.BaseURL,
		PublicClient: client.PublicClient,
		Enabled:      client.Enabled,
	}); err != nil {
		return fmt.Errorf("keycloak update failed: %w", err)
	}

	if err := r.db.UpsertClient(ctx, db.UpsertClientParams{
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
		PublicClient: sql.NullBool{
			Bool:  client.PublicClient,
			Valid: true,
		},
		Enabled: sql.NullBool{
			Bool:  client.Enabled,
			Valid: true,
		},
	}); err != nil {
		r.logger.Warn(
			"client updated in keycloak but db sync failed",
			"clientId", client.ClientID,
			"error", err,
		)
	}

	return nil
}

// DELETE (KC authoritative)
func (r *sqlcClientRepository) DeleteClient(id uuid.UUID) error {
	ctx := context.Background()

	dbClient, err := r.db.GetClientByID(ctx, id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if dbClient.ClientID != "" {
		if err := r.keycloakClient.DeleteClient(dbClient.ClientID); err != nil {
			return fmt.Errorf("failed deleting keycloak client: %w", err)
		}
	}

	if err := r.db.DeleteClient(ctx, id); err != nil {
		r.logger.Warn(
			"client deleted in keycloak but db delete failed",
			"id", id,
			"error", err,
		)
	}

	return nil
}
