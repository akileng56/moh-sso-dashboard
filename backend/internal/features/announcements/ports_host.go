package announcements

import (
	"context"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/config"
	healthcontextfeature "github.com/moh-sso-dashboard/internal/features/health_context"
	rbacfeature "github.com/moh-sso-dashboard/internal/features/rbac"
	userfeature "github.com/moh-sso-dashboard/internal/features/users"
	models "github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/storage"
)

// NewService retains the host constructor while adapting identity dependencies.
func NewService(repo Repository, users userfeature.UserRepository, notifications NotificationPublisher,
	fileStorage storage.Storage, cfg ...*config.Config) *Service {
	return NewServiceWithPorts(repo, ServicePorts{
		Users: NewUserDirectory(users), Notifications: notifications, Storage: fileStorage,
	}, cfg...)
}

type hostUserDirectory struct{ repo userfeature.UserRepository }

func NewUserDirectory(repo userfeature.UserRepository) UserDirectory {
	if repo == nil {
		return nil
	}
	return &hostUserDirectory{repo: repo}
}

func audienceUser(user models.User) AudienceUser {
	return AudienceUser{ID: user.ID, Username: user.Username, Email: user.Email,
		FirstName: user.FirstName, LastName: user.LastName, FullName: user.FullName, Enabled: user.Enabled}
}

func audienceUsers(users []models.User) []AudienceUser {
	if users == nil {
		return nil
	}
	out := make([]AudienceUser, len(users))
	for i, user := range users {
		out[i] = audienceUser(user)
	}
	return out
}

// The legacy list/detail APIs do not accept contexts. Check cancellation before
// invoking them; in-flight cancellation requires a future host API improvement.
func (a *hostUserDirectory) ListUsers(ctx context.Context) ([]AudienceUser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	users, err := a.repo.ListUsers()
	return audienceUsers(users), err
}
func (a *hostUserDirectory) GetUserByID(ctx context.Context, id uuid.UUID) (*AudienceUser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	user, err := a.repo.GetUserByID(id)
	if user == nil {
		return nil, err
	}
	out := audienceUser(*user)
	return &out, err
}
func (a *hostUserDirectory) GetUsersByRealmRole(ctx context.Context, role string) ([]AudienceUser, error) {
	users, err := a.repo.GetUsersByRealmRole(ctx, role)
	return audienceUsers(users), err
}
func (a *hostUserDirectory) GetUserClientRoles(ctx context.Context, id string) ([]ClientAccess, error) {
	assignments, err := a.repo.GetUserClientRoles(ctx, id)
	if assignments == nil {
		return nil, err
	}
	out := make([]ClientAccess, len(assignments))
	for i, assignment := range assignments {
		out[i] = ClientAccess{ID: assignment.Id, ClientID: assignment.ClientID}
	}
	return out, err
}

type hostGroupDirectory struct{ repo rbacfeature.Repository }

func NewGroupDirectory(repo rbacfeature.Repository) GroupDirectory {
	if repo == nil {
		return nil
	}
	return &hostGroupDirectory{repo: repo}
}
func (a *hostGroupDirectory) ValidateGroup(ctx context.Context, id string) error {
	_, err := a.repo.GetGroup(ctx, id)
	return err
}
func (a *hostGroupDirectory) ListGroupMembers(ctx context.Context, id string) ([]AudienceGroupMember, error) {
	members, err := a.repo.ListGroupMembers(ctx, id)
	if members == nil {
		return nil, err
	}
	out := make([]AudienceGroupMember, len(members))
	for i, member := range members {
		out[i] = AudienceGroupMember{UserID: member.UserID, Username: member.Username, Email: member.Email}
	}
	return out, err
}

// SetRBACRepository and SetHealthContextService retain existing host wiring.
func (s *Service) SetRBACRepository(repo rbacfeature.Repository) {
	if s != nil {
		s.groups = NewGroupDirectory(repo)
	}
}
func (s *Service) SetHealthContextService(service *healthcontextfeature.Service) {
	if s == nil {
		return
	}
	if service == nil {
		s.healthContexts = nil
		return
	}
	s.healthContexts = service
}
