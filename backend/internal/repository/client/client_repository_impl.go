package client

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
	models "github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/utils"
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

	return &sqlcClientRepository{
		keycloakClient: keycloakClient,
		config:         config,
		db:             db,
		logger:         &log,
	}
}

func (r *sqlcClientRepository) CreateClient(client *models.Client) (string, error) {
	ctx := context.Background()

	attributes := utils.DefaultClientAttributes(client.ClientID)

	kcID, err := r.keycloakClient.CreateClient(keycloak.CreateClientParams{
		ClientID:     client.ClientID,
		Name:         client.Name,
		Description:  client.Description,
		BaseURL:      client.BaseURL,
		PublicClient: client.PublicClient,
		Protocol:     "openid-connect",
		RedirectURIs: []string{},
		WebOrigins:   []string{},
		Attributes:   attributes,
	})
	if err != nil {
		return "", fmt.Errorf("keycloak create failed: %w", err)
	}

	kcClient, err := r.keycloakClient.GetClientByClientID(client.ClientID)
	if err != nil {
		return "", fmt.Errorf("keycloak read-back failed: %w", err)
	}

	id, _ := uuid.Parse(kcID)

	var icon sql.NullString
	if v := kcClient.Attributes["ui.icon"]; v != "" {
		icon = sql.NullString{String: v, Valid: true}
	}

	attrsJSON, err := json.Marshal(kcClient.Attributes)
	if err != nil {
		return "", fmt.Errorf("marshal client attributes failed: %w", err)
	}

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
		Icon:         icon,
		PublicClient: kcClient.PublicClient,
		Enabled:      kcClient.Enabled,
		Attributes:   attrsJSON,
	}); err != nil {
		r.logger.Warn(
			"client created in keycloak but db sync failed",
			"clientId", kcClient.ClientID,
			"error", err,
		)
	}

	return kcID, nil
}

func (r *sqlcClientRepository) GetClientByID(id uuid.UUID) (*models.Client, error) {
	ctx := context.Background()

	dbRow, err := r.db.GetClientByID(ctx, id)
	if err != nil {
		return nil, err
	}

	kcClient, err := r.keycloakClient.GetClientByClientID(dbRow.ClientID)
	if err != nil {
		return nil, err
	}

	return &models.Client{
		ID:           dbRow.ID.String(),
		ClientID:     kcClient.ClientID,
		Name:         kcClient.Name,
		Description:  kcClient.Description,
		BaseURL:      kcClient.BaseURL,
		PublicClient: kcClient.PublicClient,
		Enabled:      kcClient.Enabled,
		Attributes:   kcClient.Attributes,
	}, nil
}

func (r *sqlcClientRepository) GetClientByClientID(clientID string) (*models.Client, error) {
	kc, err := r.keycloakClient.GetClientByClientID(clientID)
	if err != nil || kc == nil {
		return nil, err
	}

	return &models.Client{
		ID:           kc.ID,
		ClientID:     kc.ClientID,
		Name:         kc.Name,
		Description:  kc.Description,
		BaseURL:      kc.BaseURL,
		PublicClient: kc.PublicClient,
		Enabled:      kc.Enabled,
		Attributes:   kc.Attributes,
	}, nil
}

func (r *sqlcClientRepository) ListClients() ([]models.Client, error) {
	kcClients, err := r.keycloakClient.ListClients()
	if err != nil {
		return nil, err
	}

	out := make([]models.Client, 0, len(kcClients))
	for _, kc := range kcClients {
		out = append(out, models.Client{
			ID:           kc.ID,
			ClientID:     kc.ClientID,
			Name:         kc.Name,
			Description:  kc.Description,
			BaseURL:      kc.BaseURL,
			PublicClient: kc.PublicClient,
			Enabled:      kc.Enabled,
			Attributes:   kc.Attributes,
		})
	}
	return out, nil
}

// UPDATE (KC first, DB best-effort)
func (r *sqlcClientRepository) UpdateClient(client *models.Client) error {
	ctx := context.Background()

	if err := r.keycloakClient.UpdateClient(client.ID, client); err != nil {
		return err
	}

	attrsJSON, _ := json.Marshal(client.Attributes)

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
		PublicClient: client.PublicClient,
		Enabled:      client.Enabled,
		Attributes:   attrsJSON,
	}); err != nil {
		r.logger.Warn(
			"client updated in keycloak but db sync failed",
			"clientId", client.ClientID,
			"error", err,
		)
	}

	return nil
}

// DELETE
func (r *sqlcClientRepository) DeleteClient(id uuid.UUID) error {
	ctx := context.Background()

	dbClient, err := r.db.GetClientByID(ctx, id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if dbClient.ClientID != "" {
		if err := r.keycloakClient.DeleteClient(dbClient.ClientID); err != nil {
			return err
		}
	}

	return r.db.DeleteClient(ctx, id)
}

func (r *sqlcClientRepository) CreateClientRole(ctx context.Context, clientID uuid.UUID, payload *models.CreateClientRoleRequest) error {
	return r.keycloakClient.CreateClientRole(
		context.Background(),
		clientID.String(),
		payload,
	)
}

func (r *sqlcClientRepository) ListClientRoles(
	ctx context.Context,
	clientID uuid.UUID,
) ([]keycloak.ClientRoleRep, error) {

	roles, err := r.keycloakClient.ListClientRoles(
		context.Background(),
		clientID.String(),
	)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *sqlcClientRepository) DeleteClientRole(ctx context.Context, clientID uuid.UUID, role string) error {
	return r.keycloakClient.DeleteClientRole(
		context.Background(),
		clientID.String(),
		role,
	)
}

func (r *sqlcClientRepository) ToggleClientEnabled(
	ctx context.Context,
	clientID uuid.UUID,
	enabled bool,
) error {

	dbClient, err := r.db.GetClientByID(ctx, clientID)
	if err != nil {
		return err
	}

	if err := r.keycloakClient.UpdateClientEnabled(ctx,
		dbClient.ClientID,
		enabled,
	); err != nil {
		return fmt.Errorf("keycloak toggle client failed: %w", err)
	}

	if err := r.db.UpdateClientEnabled(ctx, db.UpdateClientEnabledParams{
		ID:      clientID,
		Enabled: enabled,
	}); err != nil {
		r.logger.Warn(
			"client enabled toggled in keycloak but db sync failed",
			"clientId", dbClient.ClientID,
			"enabled", enabled,
			"error", err,
		)
	}

	return nil
}

func (r *sqlcClientRepository) GetClientRoleByName(
	ctx context.Context,
	clientID uuid.UUID,
	clientUuid uuid.UUID,
	role string,
) (*keycloak.ClientRoleRep, error) {

	return r.keycloakClient.GetClientRoleByName(
		ctx,
		clientID.String(),
		clientUuid.String(),
		role,
	)
}
