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

//
// -------------------------------------------------------------------
// Client lifecycle
// -------------------------------------------------------------------
//

// CREATE (Keycloak authoritative)
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

// GET by DB UUID
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

// GET by clientId
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

// LIST (Keycloak primary)
func (r *sqlcClientRepository) ListClients() ([]models.Client, error) {
	kcClients, err := r.keycloakClient.ListClients()
	if err != nil {
		return nil, err
	}

	var out []models.Client
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
		Attributes:   json.RawMessage(client.Attributes[""]),
	}); err != nil {
		r.logger.Warn("client updated in keycloak but db sync failed",
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

//
// -------------------------------------------------------------------
// Client roles / permissions
// -------------------------------------------------------------------
//

func (r *sqlcClientRepository) CreateClientRole(clientID uuid.UUID, role string) error {
	return r.keycloakClient.CreateClientRole(
		context.Background(),
		clientID.String(),
		role,
	)
}

func (r *sqlcClientRepository) ListClientRoles(clientID uuid.UUID) ([]string, error) {
	roles, err := r.keycloakClient.ListClientRoles(
		context.Background(),
		clientID.String(),
	)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(roles))
	for _, r := range roles {
		out = append(out, r.Name)
	}
	return out, nil
}

func (r *sqlcClientRepository) DeleteClientRole(clientID uuid.UUID, role string) error {
	return r.keycloakClient.DeleteClientRole(
		context.Background(),
		clientID.String(),
		role,
	)
}

//
// -------------------------------------------------------------------
// User ↔ Client role mapping
// -------------------------------------------------------------------
//

func (r *sqlcClientRepository) AssignClientRoleToUser(
	userID uuid.UUID,
	clientID uuid.UUID,
	role string,
) error {

	cr, err := r.keycloakClient.GetClientRoleByName(
		context.Background(),
		clientID.String(),
		role,
	)
	if err != nil {
		return err
	}

	return r.keycloakClient.AddClientRoleToUser(
		context.Background(),
		userID.String(),
		clientID.String(),
		*cr,
	)
}

func (r *sqlcClientRepository) RemoveClientRoleFromUser(
	userID uuid.UUID,
	clientID uuid.UUID,
	role string,
) error {

	cr, err := r.keycloakClient.GetClientRoleByName(
		context.Background(),
		clientID.String(),
		role,
	)
	if err != nil {
		return err
	}

	return r.keycloakClient.RemoveClientRoleFromUser(
		context.Background(),
		userID.String(),
		clientID.String(),
		*cr,
	)
}
