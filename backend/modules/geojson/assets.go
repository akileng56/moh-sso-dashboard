package geojson

import (
	"context"
	"io"
	"time"
)

// Asset owns a seekable stream and its modification time. The caller must
// close any non-nil returned stream, including one returned alongside an error.
type Asset struct {
	Content    io.ReadSeekCloser
	ModifiedAt time.Time
}

// AssetReader opens a canonical asset filename selected by the feature allowlist.
// It returns no HTTP, filesystem, authentication or host-specific types.
type AssetReader interface {
	Open(context.Context, string) (Asset, error)
}
