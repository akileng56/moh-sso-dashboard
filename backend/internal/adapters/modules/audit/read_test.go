package audit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type auditReadFixture struct {
	Repository
	ctx    context.Context
	inputs []ListAuditLogsInput
	items  []AuditLogResponse
	err    error
}

func (r *auditReadFixture) ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (AuditLogListResponse, error) {
	r.ctx = ctx
	r.inputs = append(r.inputs, input)
	return AuditLogListResponse{Items: r.items, HasMore: true, NextCursor: AuditCursor{CursorID: input.CursorID, CursorCreatedAt: input.CursorCreatedAt}}, r.err
}
func (r *auditReadFixture) ExportAuditLogs(context.Context, ExportAuditLogsInput) ([]AuditLogResponse, error) {
	return r.items, r.err
}

type auditCacheFixture struct {
	getCtx, setCtx context.Context
	payload        []byte
	getErr, setErr error
	gets, sets     int
	key            string
	ttl            time.Duration
}

func (c *auditCacheFixture) Get(ctx context.Context, key string, dest any) (bool, error) {
	c.getCtx = ctx
	c.gets++
	c.key = key
	if c.getErr != nil {
		return true, c.getErr
	} // A failed adapter may leave partial data.
	if c.payload == nil {
		return false, nil
	}
	if err := json.Unmarshal(c.payload, dest); err != nil {
		return false, err
	}
	return true, nil
}
func (c *auditCacheFixture) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	c.setCtx = ctx
	c.sets++
	c.key = key
	c.ttl = ttl
	if c.setErr != nil {
		return c.setErr
	}
	var err error
	c.payload, err = json.Marshal(value)
	return err
}

func auditFixtureItems() []AuditLogResponse {
	success := false
	return []AuditLogResponse{{ID: "11111111-1111-1111-1111-111111111111", CreatedAt: "2026-11-18T06:00:00Z", Username: "fixture", Action: "login_failed", IP: "127.0.0.1", ClientID: "dashboard-web", Success: &success, Metadata: map[string]any{"success": false, "message": "fixture, quoted \"value\""}}}
}

type queryCache map[string][]byte

func (c queryCache) Get(_ context.Context, key string, dest any) (bool, error) {
	payload, ok := c[key]
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal(payload, dest); err != nil {
		return false, err
	}
	return true, nil
}
func (c queryCache) Set(_ context.Context, key string, value any, _ time.Duration) error {
	payload, err := json.Marshal(value)
	if err == nil {
		c[key] = payload
	}
	return err
}

func TestCachedReaderSeparatesQueriesAndBypassesEmptyIdentity(t *testing.T) {
	repo := &auditReadFixture{items: auditFixtureItems()}
	cache := queryCache{}
	reader := NewCachedReader(NewService(repo), cache)
	for _, action := range []string{"login", "logout", "login", "logout"} {
		result, err := reader.ListAuditLogs(context.Background(), ListAuditLogsInput{Action: action}, "action="+action)
		if err != nil || len(result.Items) != 1 {
			t.Fatalf("read failed: %+v %v", result, err)
		}
	}
	if len(repo.inputs) != 2 || len(cache) != 2 || repo.inputs[0].Action != "login" || repo.inputs[1].Action != "logout" {
		t.Fatalf("query cache identities collided: calls=%d cache=%v", len(repo.inputs), cache)
	}
	for i := 0; i < 2; i++ {
		if _, err := reader.ListAuditLogs(context.Background(), ListAuditLogsInput{}, ""); err != nil {
			t.Fatal(err)
		}
	}
	if len(repo.inputs) != 4 || len(cache) != 2 {
		t.Fatal("empty cache identities must bypass caching")
	}
}

type auditContextKey struct{}

func TestCachedReaderPreservesContextAndErrors(t *testing.T) {
	ctx := context.WithValue(context.Background(), auditContextKey{}, "request-fixture")
	repo := &auditReadFixture{items: auditFixtureItems()}
	cache := &auditCacheFixture{}
	reader := NewCachedReader(NewService(repo), cache)
	if _, err := reader.ListAuditLogs(ctx, ListAuditLogsInput{}, "query"); err != nil {
		t.Fatal(err)
	}
	if repo.ctx != ctx || cache.getCtx != ctx || cache.setCtx != ctx {
		t.Fatal("request context lost at read/cache boundary")
	}
	failure := context.DeadlineExceeded
	repo.err = failure
	if _, err := reader.ListAuditLogs(ctx, ListAuditLogsInput{}, ""); !errors.Is(err, failure) {
		t.Fatalf("repository failure lost: %v", err)
	}
	for _, reader := range []*CachedReader{nil, NewCachedReader(nil, cache)} {
		if _, err := reader.ListAuditLogs(ctx, ListAuditLogsInput{}, "query"); err == nil {
			t.Fatal("missing read service accepted")
		}
	}
}
