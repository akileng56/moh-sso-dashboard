package geojson

import (
	"context"
	"errors"
)

var (
	ErrUnknownLayer = errors.New("requested geojson does not exist")
	ErrLoadFailed   = errors.New("failed to load geojson file")
)

// ResolveLayer maps public layer names to canonical filenames. It never accepts
// a path or a caller-selected filename.
func ResolveLayer(name string) (string, error) {
	switch name {
	case "geo_water":
		return "uga_water.geojson", nil
	case "geo_districts":
		return "uganda_districts.geojson", nil
	case "geo_subcounties":
		return "uganda_subcounties.geojson", nil
	default:
		return "", ErrUnknownLayer
	}
}

// Layer is a resolved asset. The caller owns its stream only on success and
// must close it after use; the service closes streams returned on failed reads.
type Layer struct {
	Filename string
	Asset
}

// Service resolves and reads layers without HTTP, filesystem or host policy.
// Callers must apply their authorization policy before invoking Open.
type Service struct {
	assets AssetReader
}

func NewService(assets AssetReader) *Service { return &Service{assets: assets} }

// loadError exposes a stable message while retaining causes for errors.Is/As.
type loadError struct{ cause error }

func (e *loadError) Error() string        { return ErrLoadFailed.Error() }
func (e *loadError) Is(target error) bool { return target == ErrLoadFailed }
func (e *loadError) Unwrap() error        { return e.cause }

// Open rejects unknown names before storage access. Failed reads return a zero
// Layer and an error matching ErrLoadFailed, with the original cause preserved.
func (s *Service) Open(ctx context.Context, name string) (Layer, error) {
	filename, err := ResolveLayer(name)
	if err != nil {
		return Layer{}, err
	}
	if err := ctx.Err(); err != nil {
		return Layer{}, &loadError{cause: err}
	}
	if s == nil || s.assets == nil {
		return Layer{}, &loadError{cause: errors.New("asset reader is required")}
	}
	asset, err := s.assets.Open(ctx, filename)
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		if asset.Content != nil {
			err = errors.Join(err, asset.Content.Close())
		}
		return Layer{}, &loadError{cause: err}
	}
	if asset.Content == nil {
		return Layer{}, &loadError{cause: errors.New("asset reader returned no content")}
	}
	return Layer{Filename: filename, Asset: asset}, nil
}
