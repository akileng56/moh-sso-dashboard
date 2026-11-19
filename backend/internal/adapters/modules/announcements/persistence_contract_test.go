package announcements

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/crc32"
	"reflect"
	"testing"
	"time"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
)

// Fill every field with distinct values, including invalid-but-nonzero nullable
// values and unknown enum strings. Converters must preserve, not normalize, them.
func fillPersistenceFixture(value any, valid bool) {
	var fill func(reflect.Value, string)
	fill = func(v reflect.Value, path string) {
		if v.Type() == reflect.TypeOf(time.Time{}) {
			v.Set(reflect.ValueOf(time.Date(2026, 11, 18, 12, 0, 0, int(crc32.ChecksumIEEE([]byte(path))%1000000000), time.FixedZone("fixture", 10800))))
			return
		}
		switch v.Kind() {
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				fill(v.Field(i), path+"."+v.Type().Field(i).Name)
			}
		case reflect.Array:
			for i := 0; i < v.Len(); i++ {
				fill(v.Index(i), fmt.Sprintf("%s.%d", path, i))
			}
		case reflect.String:
			v.SetString("fixture" + path)
		case reflect.Bool:
			v.SetBool(valid)
		case reflect.Int, reflect.Int32, reflect.Int64:
			v.SetInt(int64(crc32.ChecksumIEEE([]byte(path))%1000000) + 1)
		case reflect.Uint8:
			v.SetUint(uint64(crc32.ChecksumIEEE([]byte(path)) % 256))
		default:
			panic("unsupported fixture type: " + v.Type().String())
		}
	}
	fill(reflect.ValueOf(value).Elem(), reflect.TypeOf(value).Elem().Name())
}

func TestPersistenceContractRoundTripsEveryFieldAndNullableState(t *testing.T) {
	for _, valid := range []bool{false, true} {
		t.Run("Announcement-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.Announcement
			fillPersistenceFixture(&original, valid)
			portable := fromSQLAnnouncement(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLAnnouncement(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("AnnouncementAttachment-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.AnnouncementAttachment
			fillPersistenceFixture(&original, valid)
			portable := fromSQLAnnouncementAttachment(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLAnnouncementAttachment(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("CreateAnnouncementAttachmentParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.CreateAnnouncementAttachmentParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLCreateAnnouncementAttachmentParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLCreateAnnouncementAttachmentParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("CreateAnnouncementParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.CreateAnnouncementParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLCreateAnnouncementParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLCreateAnnouncementParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("DraftAnnouncementParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.DraftAnnouncementParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLDraftAnnouncementParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLDraftAnnouncementParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("GetAnnouncementAttachmentByIDParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.GetAnnouncementAttachmentByIDParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLGetAnnouncementAttachmentByIDParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLGetAnnouncementAttachmentByIDParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("GetAnnouncementStatsRow-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.GetAnnouncementStatsRow
			fillPersistenceFixture(&original, valid)
			portable := fromSQLGetAnnouncementStatsRow(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLGetAnnouncementStatsRow(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("ListActivePublishedAnnouncementsParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.ListActivePublishedAnnouncementsParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLListActivePublishedAnnouncementsParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLListActivePublishedAnnouncementsParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("ListAnnouncementsAdminParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.ListAnnouncementsAdminParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLListAnnouncementsAdminParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLListAnnouncementsAdminParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("ListAnnouncementsByStatusParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.ListAnnouncementsByStatusParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLListAnnouncementsByStatusParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLListAnnouncementsByStatusParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("ListAnnouncementsCreatedByUserParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.ListAnnouncementsCreatedByUserParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLListAnnouncementsCreatedByUserParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLListAnnouncementsCreatedByUserParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("ListAnnouncementsForClientParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.ListAnnouncementsForClientParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLListAnnouncementsForClientParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLListAnnouncementsForClientParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("ListAnnouncementsForRoleParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.ListAnnouncementsForRoleParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLListAnnouncementsForRoleParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLListAnnouncementsForRoleParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("ListAnnouncementsForUserParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.ListAnnouncementsForUserParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLListAnnouncementsForUserParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLListAnnouncementsForUserParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("ScheduleAnnouncementParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.ScheduleAnnouncementParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLScheduleAnnouncementParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLScheduleAnnouncementParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("SearchAnnouncementsAdminParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.SearchAnnouncementsAdminParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLSearchAnnouncementsAdminParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLSearchAnnouncementsAdminParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("SoftDeleteAnnouncementAttachmentParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.SoftDeleteAnnouncementAttachmentParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLSoftDeleteAnnouncementAttachmentParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLSoftDeleteAnnouncementAttachmentParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("UpdateAnnouncementAttachmentParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.UpdateAnnouncementAttachmentParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLUpdateAnnouncementAttachmentParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLUpdateAnnouncementAttachmentParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("UpdateAnnouncementParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.UpdateAnnouncementParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLUpdateAnnouncementParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLUpdateAnnouncementParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
		t.Run("UpdateAnnouncementStatusParams-"+fmt.Sprint(valid), func(t *testing.T) {
			var original db.UpdateAnnouncementStatusParams
			fillPersistenceFixture(&original, valid)
			portable := fromSQLUpdateAnnouncementStatusParams(original)
			if reflect.TypeOf(portable).NumField() != reflect.TypeOf(original).NumField() {
				t.Fatal("persistence field coverage drift")
			}
			if got := toSQLUpdateAnnouncementStatusParams(portable); !reflect.DeepEqual(got, original) {
				t.Fatalf("round trip lost data: got=%+v want=%+v", got, original)
			}
		})
	}
}

type sqlcRepositoryFixture struct {
	SQLCRepository
	ctx    context.Context
	params db.CreateAnnouncementParams
	item   db.Announcement
	items  []db.Announcement
	status model.AnnouncementStatus
	err    error
}

func (f *sqlcRepositoryFixture) Create(ctx context.Context, params db.CreateAnnouncementParams) (db.Announcement, error) {
	f.ctx, f.params = ctx, params
	return f.item, f.err
}
func (f *sqlcRepositoryFixture) ListAdmin(ctx context.Context, _ db.ListAnnouncementsAdminParams) ([]db.Announcement, error) {
	f.ctx = ctx
	return f.items, f.err
}
func (f *sqlcRepositoryFixture) CountByStatus(ctx context.Context, status model.AnnouncementStatus) (int64, error) {
	f.ctx, f.status = ctx, status
	return 7, f.err
}

func TestPersistenceAdapterPreservesQueriesContextErrorsAndCollectionShape(t *testing.T) {
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "request")
	raw := &sqlcRepositoryFixture{}
	fillPersistenceFixture(&raw.item, true)
	var params CreateAnnouncementParams
	fillPersistenceFixture(&params, true)
	adapter := &persistenceAdapter{host: raw}
	item, err := adapter.Create(ctx, params)
	if err != nil || raw.ctx != ctx || !reflect.DeepEqual(raw.params, toSQLCreateAnnouncementParams(params)) ||
		!reflect.DeepEqual(item, fromSQLAnnouncement(raw.item)) {
		t.Fatalf("create boundary lost data: %v", err)
	}
	for _, items := range [][]db.Announcement{nil, {}, {raw.item}} {
		raw.items = items
		result, err := adapter.ListAdmin(ctx, ListAnnouncementsAdminParams{Limit: 10, Offset: 2})
		if err != nil || (result == nil) != (items == nil) || len(result) != len(items) {
			t.Fatalf("collection shape changed: %+v %v", result, err)
		}
	}
	count, err := adapter.CountByStatus(ctx, AnnouncementStatusPUBLISHED)
	if count != 7 || err != nil || raw.status != model.AnnouncementStatusPUBLISHED {
		t.Fatal("status conversion changed")
	}
	raw.err = sql.ErrNoRows
	if _, err := adapter.Create(ctx, params); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("persistence cause lost: %v", err)
	}
}
