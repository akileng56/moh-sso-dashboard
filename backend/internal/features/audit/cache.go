package audit

import (
	"context"
	"time"
)

// Cache is an optional best-effort read cache. Get decodes into dest; errors are
// treated as misses and Set failures never turn successful queries into errors.
// Authentication and audit permissions are enforced by the host before lookup.
type Cache interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}
