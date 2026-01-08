package repository

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
	models "github.com/moh-sso-dashboard/internal/model"
)

type userRepository struct {
	keycloakClient *keycloak.Client
	config         *config.Config
	db             db.Store
	logger         *logger.Logger
}

func NewUserRepository(
	keycloakClient *keycloak.Client,
	config *config.Config,
	store db.Store,
	log logger.Logger,
) UserRepository {

	// Authenticate Keycloak admin
	if err := keycloakClient.Authenticate(); err != nil {
		panic(fmt.Sprintf("❌ Failed to authenticate Keycloak admin: %v", err))
	}

	return &userRepository{
		keycloakClient: keycloakClient,
		config:         config,
		db:             store,
		logger:         &log,
	}
}

// ------------------------------------------------------------
// Create user (Keycloak → DB)
// ------------------------------------------------------------
func (r *userRepository) CreateUser(user *models.User) (string, error) {
	ctx := context.Background()

	kcID, err := r.keycloakClient.CreateUser(user)
	if err != nil {
		r.logger.Error("Failed creating user in Keycloak: %v", err)
		return "", fmt.Errorf("keycloak user creation failed: %w", err)
	}

	userID, err := uuid.Parse(kcID)

	if err != nil {
		r.logger.Error("Failed to parse Keycloak ID: %v", err)
		return "", fmt.Errorf("failed to parse Keycloak ID %v", err)
	}

	params := db.CreateUserParams{
		ID:        userID,
		Username:  user.Username,
		FirstName: sql.NullString{String: user.FirstName, Valid: user.FirstName != ""},
		LastName:  sql.NullString{String: user.LastName, Valid: user.LastName != ""},
		Email:     user.Email,
		Enabled: sql.NullBool{
			Bool:  user.Enabled,
			Valid: user.Enabled,
		},
	}

	if err := r.db.CreateUser(ctx, params); err != nil {
		r.logger.Error("Failed inserting user into DB: %v", err)
		return "", err
	}

	return kcID, nil
}

// ------------------------------------------------------------
// Get user by ID
// ------------------------------------------------------------
func (r *userRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	// 1. Fetch from Keycloak (source of truth)
	kcUser, err := r.keycloakClient.GetUser(id.String())
	if err != nil {
		r.logger.Error("Failed creating user in Keycloak: %v", err)
		return nil, fmt.Errorf("keycloak user creation failed: %w", err)
	}

	// 2. Optional: enrich from DB (local metadata)
	ctx := context.Background()
	row, err := r.db.GetUserByID(ctx, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	var createdAt time.Time
	var updatedAt time.Time

	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}

	// 3. Build unified user model
	return &models.User{
		ID:            kcUser.ID,
		Username:      kcUser.Username,
		Email:         kcUser.Email,
		FirstName:     kcUser.FirstName,
		LastName:      kcUser.LastName,
		FullName:      strings.TrimSpace(kcUser.FirstName + " " + kcUser.LastName),
		RealmRoles:    kcUser.Roles,
		ClientRoles:   kcUser.ClientRoles,
		IsAdmin:       slices.Contains(kcUser.Roles, "admin"),
		Enabled:       kcUser.Enabled,
		EmailVerified: kcUser.EmailVerified,
		LastLoginAt:   kcUser.LastLoginAt,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

// ------------------------------------------------------------
// List all users
// ------------------------------------------------------------
func (r *userRepository) ListUsers() ([]models.User, error) {
	kcUsers, err := r.keycloakClient.ListUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch keycloak users: %w", err)
	}

	result := make([]models.User, 0, len(kcUsers))

	for _, kc := range kcUsers {
		result = append(result, models.User{
			ID:            kc.ID,
			Username:      kc.Username,
			Email:         kc.Email,
			FirstName:     kc.FirstName,
			LastName:      kc.LastName,
			FullName:      strings.TrimSpace(kc.FirstName + " " + kc.LastName),
			RealmRoles:    kc.Roles,
			ClientRoles:   kc.ClientRoles,
			IsAdmin:       slices.Contains(kc.Roles, "admin"),
			Enabled:       kc.Enabled,
			EmailVerified: kc.EmailVerified,
			LastLoginAt:   kc.LastLoginAt,
			CreatedAt:     kc.CreatedAt,
		})
	}

	return result, nil
}

// ------------------------------------------------------------
// Update user (Keycloak → DB)
// ------------------------------------------------------------
func (r *userRepository) UpdateUser(user *models.User) error {
	ctx := context.Background()

	// ----------------------------------------------------
	// 1️⃣ Update user in Keycloak (source of truth)
	// ----------------------------------------------------
	if err := r.keycloakClient.UpdateUser(&model.User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Enabled:   user.Enabled,
	}); err != nil {
		r.logger.Error("Failed to update user in Keycloak: %v", err)
		return fmt.Errorf("keycloak update failed: %w", err)
	}

	// ----------------------------------------------------
	// 2️⃣ Update local metadata in Postgres
	// ----------------------------------------------------
	params := db.UpdateUserParams{
		ID: uuid.MustParse(user.ID),

		FirstName: sql.NullString{
			String: user.FirstName,
			Valid:  user.FirstName != "",
		},

		LastName: sql.NullString{
			String: user.LastName,
			Valid:  user.LastName != "",
		},

		Email: user.Email,
	}

	if err := r.db.UpdateUser(ctx, params); err != nil {
		r.logger.Error("Failed to update user in DB: %v", err)
		return fmt.Errorf("db update failed: %w", err)
	}

	return nil
}

// ------------------------------------------------------------
// Delete User (Keycloak → DB)
// ------------------------------------------------------------
func (r *userRepository) DeleteUser(id string) error {
	ctx := context.Background()

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}
	if err := r.keycloakClient.DeleteUser(id); err != nil {
		r.logger.Error("Keycloak failed to delete user %s: %v", id, err)
		return fmt.Errorf("keycloak delete failed: %w", err)
	}
	if err := r.db.DeleteUser(ctx, uid); err != nil {
		r.logger.Error("DB failed to delete user %s: %v", id, err)
		return fmt.Errorf("db delete failed: %w", err)
	}

	return nil
}
