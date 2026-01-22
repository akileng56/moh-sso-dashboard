package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/keycloak"
	models "github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/user"
	"github.com/moh-sso-dashboard/internal/utils"
)

type CreateUserRequest struct {
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	FirstName        string              `json:"first_name,omitempty"`
	LastName         string              `json:"last_name,omitempty"`
	FullName         string              `json:"full_name,omitempty"`
	Enabled          *bool               `json:"enabled,omitempty"`
	RequirePwdChange *bool               `json:"require_pwd_change,omitempty"`
	Password         string              `json:"password,omitempty"`
	SendInvite       bool                `json:"send_invite,omitempty"`
	RealmRoles       []string            `json:"realm_roles,omitempty"`
	ClientRoles      map[string][]string `json:"client_roles,omitempty"`
}

type UserService struct {
	repo          repository.UserRepository
	notifications NotificationsService
}

func NewUserService(
	repo repository.UserRepository,
	notifications NotificationsService,
) *UserService {
	return &UserService{
		repo:          repo,
		notifications: notifications,
	}
}

//
// ----------------------------------------------------
// USER LIFECYCLE
// ----------------------------------------------------
//

func (s *UserService) CreateUser(
	ctx context.Context,
	req CreateUserRequest,
	adminID uuid.UUID,
) (*models.User, error) {

	if req.Username == "" || req.Email == "" {
		return nil, errors.New("username and email are required")
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Enabled:   enabled,
	}

	if _, err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	nt := models.UserCreated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "New user account created",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
			"admin_id": adminID.String(),
		}),
	})

	return user, nil
}

func (s *UserService) SetUserEnabled(
	ctx context.Context,
	userID uuid.UUID,
	enabled bool,
	adminID uuid.UUID,
) error {

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user.Enabled == enabled {
		return nil
	}

	if err := s.repo.ToggleUserEnabled(ctx, userID.String(), enabled); err != nil {
		return err
	}

	var nt models.NotificationType
	msg := "User account updated"

	if enabled {
		nt = models.UserEnabled
		msg = "User account enabled"
	} else {
		nt = models.UserDisabled
		msg = "User account disabled"
	}

	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    msg,
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"user_id":  user.ID,
			"username": user.Username,
			"admin_id": adminID.String(),
		}),
	})

	return nil
}

func (s *UserService) ResetUserPassword(
	ctx context.Context,
	userID uuid.UUID,
	adminID uuid.UUID,
) error {

	if err := s.repo.ResetUserPassword(ctx, userID.String()); err != nil {
		return err
	}

	nt := models.UserPasswordReset
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "User password reset",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"user_id":  userID.String(),
			"admin_id": adminID.String(),
		}),
	})

	return nil
}

func (s *UserService) GetUserClientRoles(
	ctx context.Context,
	userID uuid.UUID,
) ([]keycloak.UserClientRoleAssignment, error) {

	return s.repo.GetUserClientRoles(ctx, userID.String())
}
func (s *UserService) GetUserClientRolesForClient(
	ctx context.Context,
	userID uuid.UUID,
	clientID uuid.UUID,
) ([]keycloak.ClientRoleRep, error) {

	return s.repo.GetUserClientRolesForClient(
		ctx,
		userID.String(),
		clientID.String(),
	)
}

func (s *UserService) UpdateUserClientRoles(
	ctx context.Context,
	userID uuid.UUID,
	clientID uuid.UUID,
	roles []string,
	adminID uuid.UUID,
) error {

	current, err := s.repo.GetUserClientRolesForClient(
		ctx,
		userID.String(),
		clientID.String(),
	)
	if err != nil {
		return err
	}

	currentSet := make(map[string]bool)
	for _, r := range current {
		currentSet[r.Name] = true
	}

	desiredSet := make(map[string]bool)
	for _, r := range roles {
		desiredSet[r] = true
	}

	var toAdd, toRemove []string

	for r := range desiredSet {
		if !currentSet[r] {
			toAdd = append(toAdd, r)
		}
	}

	for r := range currentSet {
		if !desiredSet[r] {
			toRemove = append(toRemove, r)
		}
	}

	if len(toAdd) > 0 {
		if err := s.repo.AddUserClientRoles(
			ctx,
			userID.String(),
			clientID.String(),
			toAdd,
		); err != nil {
			return err
		}
	}

	if len(toRemove) > 0 {
		if err := s.repo.RemoveUserClientRoles(
			ctx,
			userID.String(),
			clientID.String(),
			toRemove,
		); err != nil {
			return err
		}
	}

	nt := models.ClientRolesUpdated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "User client roles updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"user_id":   userID.String(),
			"client_id": clientID.String(),
			"added":     toAdd,
			"removed":   toRemove,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}

func (s *UserService) GetUser(id uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *UserService) ListUsers() ([]models.User, error) {
	return s.repo.ListUsers()
}

func (s *UserService) DeleteUser(
	ctx context.Context,
	id uuid.UUID,
	adminID uuid.UUID,
) error {

	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteUser(id.String()); err != nil {
		return err
	}

	nt := models.UserDeleted
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "User account deleted",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"user_id":  user.ID,
			"username": user.Username,
			"admin_id": adminID.String(),
		}),
	})

	return nil
}
