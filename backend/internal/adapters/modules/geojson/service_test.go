package geojson

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
	"time"
)

type portableReader func(context.Context, string) (Asset, error)

func (f portableReader) Open(ctx context.Context, filename string) (Asset, error) {
	return f(ctx, filename)
}

type portableStream struct {
	*bytes.Reader
	closed   int
	closeErr error
}

func (s *portableStream) Close() error { s.closed++; return s.closeErr }

type readerFailure struct{ path string }

func (e *readerFailure) Error() string { return e.path }

func TestServiceResolvesPublicLayersAndTransfersSuccessfulStream(t *testing.T) {
	for name, filename := range map[string]string{
		"geo_water":       "uga_water.geojson",
		"geo_districts":   "uganda_districts.geojson",
		"geo_subcounties": "uganda_subcounties.geojson",
	} {
		t.Run(name, func(t *testing.T) {
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "request")
			stream := &portableStream{Reader: bytes.NewReader([]byte("{}"))}
			modified := time.Date(2026, 11, 1, 12, 0, 0, 123, time.UTC)
			calls := 0
			svc := NewService(portableReader(func(got context.Context, actual string) (Asset, error) {
				calls++
				if got != ctx || actual != filename {
					t.Fatalf("context or canonical filename changed: %q", actual)
				}
				return Asset{Content: stream, ModifiedAt: modified}, nil
			}))
			layer, err := svc.Open(ctx, name)
			if err != nil {
				t.Fatal(err)
			}
			if calls != 1 || layer.Filename != filename || layer.Content != stream ||
				!layer.ModifiedAt.Equal(modified) || stream.closed != 0 {
				t.Fatalf("invalid ownership or asset: %+v", layer)
			}
			body, err := io.ReadAll(layer.Content)
			if err != nil || string(body) != "{}" {
				t.Fatalf("body=%q err=%v", body, err)
			}
			_ = layer.Content.Close()
			if stream.closed != 1 {
				t.Fatal("caller could not close stream")
			}
		})
	}
}

func TestServiceRejectsUnknownNamesBeforeAccessingStorage(t *testing.T) {
	calls := 0
	svc := NewService(portableReader(func(context.Context, string) (Asset, error) {
		calls++
		return Asset{}, errors.New("must not open")
	}))
	for _, name := range []string{"", "unknown", "../secret", "geo_water.geojson", "geo_water/..", "GEO_WATER", "geo_water\x00"} {
		filename, err := ResolveLayer(name)
		if filename != "" || !errors.Is(err, ErrUnknownLayer) {
			t.Fatalf("invalid resolution for %q", name)
		}
		layer, err := svc.Open(context.Background(), name)
		if !errors.Is(err, ErrUnknownLayer) || errors.Is(err, ErrLoadFailed) || layer.Content != nil {
			t.Fatalf("unknown name %q: layer=%+v error=%v", name, layer, err)
		}
	}
	if calls != 0 {
		t.Fatalf("storage opened %d times", calls)
	}
}

func TestServiceClassifiesLoadErrorsWithoutLeakingPaths(t *testing.T) {
	private := &readerFailure{path: "/private/maps/uga_water.geojson"}
	closeFailure := errors.New("close failed")
	for _, cause := range []error{private, fs.ErrNotExist} {
		stream := &portableStream{Reader: bytes.NewReader(nil), closeErr: closeFailure}
		svc := NewService(portableReader(func(context.Context, string) (Asset, error) {
			return Asset{Content: stream}, cause
		}))
		layer, err := svc.Open(context.Background(), "geo_water")
		if !errors.Is(err, ErrLoadFailed) || !errors.Is(err, cause) ||
			!errors.Is(err, closeFailure) || errors.Is(err, ErrUnknownLayer) {
			t.Fatalf("error classification lost: %v", err)
		}
		if err.Error() != ErrLoadFailed.Error() || strings.Contains(err.Error(), "/private") {
			t.Fatalf("unsafe error message: %v", err)
		}
		if cause == private {
			var found *readerFailure
			if !errors.As(err, &found) || found != private {
				t.Fatal("typed reader cause lost")
			}
		}
		if layer != (Layer{}) || stream.closed != 1 {
			t.Fatalf("failed stream ownership: layer=%+v closes=%d", layer, stream.closed)
		}
	}
}

func TestServiceMissingReadersAndEmptyAssetsFailSafely(t *testing.T) {
	var missing *Service
	for _, svc := range []*Service{missing, {}, NewService(nil), NewService(portableReader(func(context.Context, string) (Asset, error) {
		return Asset{}, nil
	}))} {
		layer, err := svc.Open(context.Background(), "geo_water")
		if !errors.Is(err, ErrLoadFailed) || layer != (Layer{}) {
			t.Fatalf("expected safe load failure: layer=%+v error=%v", layer, err)
		}
		if _, err := svc.Open(context.Background(), "unknown"); !errors.Is(err, ErrUnknownLayer) {
			t.Fatalf("unknown layer precedence changed: %v", err)
		}
	}
}

func TestServiceCancellationPreservesCauseAndClosesPartialAsset(t *testing.T) {
	t.Run("before-read", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		calls := 0
		svc := NewService(portableReader(func(context.Context, string) (Asset, error) {
			calls++
			return Asset{}, nil
		}))
		layer, err := svc.Open(ctx, "geo_water")
		if calls != 0 || layer != (Layer{}) || !errors.Is(err, context.Canceled) || !errors.Is(err, ErrLoadFailed) {
			t.Fatalf("cancellation lost: %v", err)
		}
	})
	t.Run("during-read", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		stream := &portableStream{Reader: bytes.NewReader(nil)}
		svc := NewService(portableReader(func(context.Context, string) (Asset, error) {
			cancel()
			return Asset{Content: stream}, nil
		}))
		layer, err := svc.Open(ctx, "geo_water")
		if layer != (Layer{}) || stream.closed != 1 || !errors.Is(err, context.Canceled) || !errors.Is(err, ErrLoadFailed) {
			t.Fatalf("canceled read retained stream: %v", err)
		}
	})
}
