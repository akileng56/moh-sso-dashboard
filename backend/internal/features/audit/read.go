package audit

import (
	"context"
	"errors"
	"time"
)

// CachedReader serves typed audit queries using optional, best-effort caching.
// The caller owns query validation and authorization before invoking the reader.
type CachedReader struct {
	service Service
	cache   Cache
}

func NewCachedReader(service Service, cache Cache) *CachedReader {
	return &CachedReader{service: service, cache: cache}
}

// ListAuditLogs uses a caller-supplied identity covering the complete query,
// including its filters and cursor. Empty identities bypass caching. The HTTP
// adapter supplies the existing raw-query identity to preserve its cache keys.
func (r *CachedReader) ListAuditLogs(ctx context.Context, input ListAuditLogsInput, identity string) (AuditLogListResponse, error) {
	if r == nil || r.service == nil {
		return AuditLogListResponse{}, errors.New("audit read service is required")
	}
	key := "audit_logs:" + identity
	useCache := r.cache != nil && identity != ""
	if useCache {
		var cached AuditLogListResponse
		if hit, err := r.cache.Get(ctx, key, &cached); err == nil && hit {
			return cached, nil
		}
	}
	result, err := r.service.ListAuditLogs(ctx, input)
	if err != nil {
		return AuditLogListResponse{}, err
	}
	if useCache {
		_ = r.cache.Set(ctx, key, result, 20*time.Second)
	}
	return result, nil
}
