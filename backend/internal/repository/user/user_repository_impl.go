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

// CREATE (Keycloak = Source of Truth)
func (r *userRepository) CreateUser(user *models.User) (string, error) {
	ctx := context.Background()

	// 1️⃣ Create user in Keycloak (AUTHORITATIVE)
	kcID, err := r.keycloakClient.CreateUser(user)
	if err != nil {
		r.logger.Error("failed creating user in keycloak", "error", err)
		return "", fmt.Errorf("keycloak user creation failed: %w", err)
	}

	uid, err := uuid.Parse(kcID)
	if err != nil {
		return "", fmt.Errorf("invalid keycloak user id: %w", err)
	}

	// 2️⃣ Best-effort DB insert (cache only)
	if err := r.db.UpsertUser(ctx, db.UpsertUserParams{
		ID:       uid,
		Username: user.Username,
		Email:    user.Email,
		FirstName: sql.NullString{
			String: user.FirstName,
			Valid:  user.FirstName != "",
		},
		LastName: sql.NullString{
			String: user.LastName,
			Valid:  user.LastName != "",
		},
		Enabled: sql.NullBool{
			Bool:  user.Enabled,
			Valid: true,
		},
	}); err != nil {
		r.logger.Warn(
			"user created in keycloak but db sync failed",
			"userId", kcID,
			"error", err,
		)
	}

	return kcID, nil
}

// GET BY ID (Keycloak authoritative)
func (r *userRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	ctx := context.Background()

	// 1️⃣ Always read from Keycloak
	kcUser, err := r.keycloakClient.GetUser(id.String())
	if err != nil {
		return nil, fmt.Errorf("keycloak user not found: %w", err)
	}

	// 2️⃣ Optional DB metadata
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

// LIST (Keycloak primary)
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

// UPDATE (Keycloak first)
func (r *userRepository) UpdateUser(user *models.User) error {
	ctx := context.Background()

	// 1️⃣ Update in Keycloak
	if err := r.keycloakClient.UpdateUser(&models.User{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Enabled:   user.Enabled,
	}); err != nil {
		return fmt.Errorf("keycloak update failed: %w", err)
	}

	// 2️⃣ Best-effort DB sync
	if err := r.db.UpsertUser(ctx, db.UpsertUserParams{
		ID:       uuid.MustParse(user.ID),
		Username: user.Username,
		Email:    user.Email,
		FirstName: sql.NullString{
			String: user.FirstName,
			Valid:  user.FirstName != "",
		},
		LastName: sql.NullString{
			String: user.LastName,
			Valid:  user.LastName != "",
		},
		Enabled: sql.NullBool{
			Bool:  user.Enabled,
			Valid: true,
		},
	}); err != nil {
		r.logger.Warn(
			"user updated in keycloak but db sync failed",
			"userId", user.ID,
			"error", err,
		)
	}

	return nil
}

// DELETE (Keycloak authoritative)
func (r *userRepository) DeleteUser(id string) error {
	ctx := context.Background()

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	// 1️⃣ Delete in Keycloak
	if err := r.keycloakClient.DeleteUser(id); err != nil {
		return fmt.Errorf("keycloak delete failed: %w", err)
	}

	// 2️⃣ Best-effort DB delete
	if err := r.db.DeleteUser(ctx, uid); err != nil {
		r.logger.Warn(
			"user deleted in keycloak but db delete failed",
			"userId", id,
			"error", err,
		)
	}

	return nil
}
