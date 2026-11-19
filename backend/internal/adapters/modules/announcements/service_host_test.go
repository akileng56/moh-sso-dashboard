package announcements

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	announcementmodule "github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements"
)

type hostCoreRepository struct {
	Repository
	ctx     context.Context
	params  CreateAnnouncementParams
	failure error
}

func (r *hostCoreRepository) Create(ctx context.Context, p CreateAnnouncementParams) (Announcement, error) {
	r.ctx, r.params = ctx, p
	return Announcement{ID: uuid.New(), Title: p.Title, Level: p.Level, Status: p.Status}, r.failure
}
func TestHostCompositionUsesPinnedCoreAndCompatibilityPorts(t *testing.T) {
	if reflect.TypeOf(Announcement{}).PkgPath() != announcementmodule.ModulePath {
		t.Fatal("host model is a retained implementation")
	}
	repo := &hostCoreRepository{}
	host := NewServiceWithPorts(repo, ServicePorts{})
	var core *announcementmodule.Service = host.Service
	if core == nil {
		t.Fatal("missing selected core")
	}
	ctx := context.Background()
	actor := uuid.New()
	item, err := host.CreateAnnouncementFromInput(ctx, CreateAnnouncementInput{
		Title: "Host adapter", Message: "Message", Level: "warning", Status: "draft", AudienceType: "ALL_USERS", CreatedBy: actor,
	})
	if err != nil || item.Level != AnnouncementLevelWARNING || repo.ctx != ctx || repo.params.CreatedBy != actor {
		t.Fatalf("host did not delegate to package: %+v %v", item, err)
	}
	groups := &legacyGroupsFixture{}
	host.SetRBACRepository(groups)
	if err := host.ValidateGroupAudience(ctx, []uuid.UUID{uuid.New()}); err != nil || groups.ctx != ctx {
		t.Fatalf("compatibility group wiring lost: %v", err)
	}
	host.SetRBACRepository(nil)
	if err := host.ValidateGroupAudience(ctx, []uuid.UUID{uuid.New()}); err == nil {
		t.Fatal("cleared group adapter remained active")
	}
	host.SetHealthContextService(nil)
	cause := errors.New("database unavailable")
	repo.failure = cause
	if _, err := host.CreateAnnouncement(ctx, CreateAnnouncementParams{}); !errors.Is(err, cause) {
		t.Fatalf("recomposition lost repository error: %v", err)
	}
}
