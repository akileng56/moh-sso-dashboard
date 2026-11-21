package geojson_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	geojson "github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson"
)

type consumerStream struct {
	*bytes.Reader
	closed int
}

func (s *consumerStream) Close() error { s.closed++; return nil }

type consumerReader struct {
	ctx      context.Context
	filename string
	stream   *consumerStream
	err      error
	calls    int
}

var _ geojson.AssetReader = (*consumerReader)(nil)

func (r *consumerReader) Open(ctx context.Context, name string) (geojson.Asset, error) {
	r.calls++
	r.ctx = ctx
	r.filename = name
	return geojson.Asset{Content: r.stream, ModifiedAt: time.Unix(100, 0).UTC()}, r.err
}

func TestExternalConsumerOwnsStorageAndSuccessfulStream(t *testing.T) {
	stream := &consumerStream{Reader: bytes.NewReader([]byte("{}"))}
	reader := &consumerReader{stream: stream}
	service := geojson.NewService(reader)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	layer, err := service.Open(ctx, "geo_districts")
	if err != nil {
		t.Fatal(err)
	}
	if reader.ctx != ctx || reader.filename != "uganda_districts.geojson" ||
		layer.Filename != reader.filename || !layer.ModifiedAt.Equal(time.Unix(100, 0)) ||
		stream.closed != 0 {
		t.Fatal("public context, asset or ownership contract changed")
	}
	data, err := io.ReadAll(layer.Content)
	if err != nil || string(data) != "{}" {
		t.Fatalf("public stream failed: %q %v", data, err)
	}
	if err := layer.Content.Close(); err != nil {
		t.Fatal(err)
	}
	if stream.closed != 1 {
		t.Fatal("caller did not own successful stream")
	}
	if _, err := service.Open(ctx, "../secret"); !errors.Is(err, geojson.ErrUnknownLayer) || reader.calls != 1 {
		t.Fatal("unknown layer reached storage")
	}
}

func TestExternalConsumerCanClassifySafeLoadErrorsAndCauses(t *testing.T) {
	cause := errors.New("/private/maps/internal.geojson")
	stream := &consumerStream{Reader: bytes.NewReader(nil)}
	reader := &consumerReader{stream: stream, err: cause}
	layer, err := geojson.NewService(reader).Open(context.Background(), "geo_water")
	if !errors.Is(err, geojson.ErrLoadFailed) || !errors.Is(err, cause) ||
		err.Error() != geojson.ErrLoadFailed.Error() || layer.Content != nil || stream.closed != 1 {
		t.Fatalf("public failure contract changed: layer=%+v error=%v", layer, err)
	}
	if geojson.ModulePath != "github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson" || geojson.Version() == "" {
		t.Fatal("public version contract missing")
	}
}

func ExampleResolveLayer() {
	filename, err := geojson.ResolveLayer("geo_water")
	fmt.Println(filename, err)
	// Output: uga_water.geojson <nil>
}
