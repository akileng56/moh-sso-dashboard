package client

import (
	"context"
	"database/sql"
	"fmt"
	"log"

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
func (r *sqlcClientRepository) CreateClient(client *models.Client) (string, error) {
	ctx := context.Background()

	kcID, err := r.keycloakClient.CreateClient(keycloak.CreateClientParams{
		ClientID:     client.ClientID,
		Name:         client.Name,
		Description:  client.Description,
		BaseURL:      client.BaseURL,
		PublicClient: client.PublicClient,
		Protocol:     "openid-connect",
		RedirectURIs: []string{},
		WebOrigins:   []string{},
		Attributes: map[string]string{
			"icon": "applications",
		}})
	if err != nil {
		return "", fmt.Errorf("keycloak create failed: %w", err)
	}

	kcClient, err := r.keycloakClient.GetClientByClientID(client.ClientID)
	if err != nil {
		return "", fmt.Errorf("keycloak read-back failed: %w", err)
	}

	var icon sql.NullString
	if kcClient.Attributes != nil {
		if v, ok := kcClient.Attributes["icon"]; ok && v != "" {
			icon = sql.NullString{
				String: v,
				Valid:  true,
			}
		}
	}

	id, _ := uuid.Parse(kcID)
	if err := r.db.UpsertClient(ctx, db.UpsertClientParams{
		ID:       id,
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
		Icon: icon,
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
	return kcID, nil
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

	log.Printf("kcClients --> %+v", kcClients)

	dbClients, err := r.db.ListClients(ctx)
	if err != nil {
		return nil, err
	}

	// DB indexed strictly by DB ID
	dbMap := map[string]db.Client{}
	for _, c := range dbClients {
		dbMap[c.ID.String()] = c
	}

	var result []models.Client

	for _, kc := range kcClients {
		attributes := map[string]string{}
		if icon, ok := kc.Attributes["icon"]; ok {
			attributes["icon"] = icon
		}

		log.Printf("ids", kc.ID)

		// 🔑 LOOKUP BY KEYCLOAK ID
		if _, ok := dbMap[kc.ID]; ok {
			// Exists in DB
			result = append(result, models.Client{
				ID:           kc.ID, // Keycloak internal ID
				ClientID:     kc.ClientID,
				Name:         kc.Name,
				Description:  kc.Description,
				BaseURL:      kc.BaseURL,
				PublicClient: kc.PublicClient,
				Enabled:      kc.Enabled,
				Attributes:   attributes,
			})
		} else {
			// Exists only in Keycloak
			result = append(result, models.Client{
				ID:           kc.ID, // still returned
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

	log.Printf("result-3 --> %+v", result)

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
