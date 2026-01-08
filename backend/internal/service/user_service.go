package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

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

// ----------------------------------------------------
// CREATE USER
// ----------------------------------------------------
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

	// 🔔 Notification: User Created
	nt := models.UserCreated
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "New user account created",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
			"admin_id": adminID.String(),
		}),
	})

	return user, nil
}

// ----------------------------------------------------
// ENABLE / DISABLE USER (USES UpdateUser)
// ----------------------------------------------------
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
		return nil // no-op
	}

	user.Enabled = enabled

	if err := s.repo.UpdateUser(user); err != nil {
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
		Metadata: utils.MustJSON(map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"admin_id": adminID.String(),
		}),
	})

	return nil
}

// ----------------------------------------------------
// CHANGE USER ROLE (CRITICAL)
// ----------------------------------------------------
func (s *UserService) ChangeUserRole(
	ctx context.Context,
	userID uuid.UUID,
	newRoles []string,
	adminID uuid.UUID,
) error {

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return err
	}

	oldRoles := user.RealmRoles
	user.RealmRoles = newRoles

	if err := s.repo.UpdateUser(user); err != nil {
		return err
	}

	nt := models.UserRoleChanged
	s.notifications.Notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(), // critical
		Message:    "User roles changed",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]interface{}{
			"user_id":   user.ID,
			"old_roles": oldRoles,
			"new_roles": newRoles,
			"admin_id":  adminID.String(),
		}),
	})

	return nil
}

// ----------------------------------------------------
// GET / LIST / DELETE
// ----------------------------------------------------
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
		Metadata: utils.MustJSON(map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"admin_id": adminID.String(),
		}),
	})

	return nil
}
