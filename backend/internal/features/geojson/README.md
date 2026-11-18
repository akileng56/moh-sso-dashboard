# GeoJSON extraction

The existing `geojson` registry module serves permission-guarded map assets.
It remains hosted in the dashboard and reports the host version.
`DISABLED_MODULES=geojson` removes its routes on restart. It owns no workers or
database migrations.

## Phase 1: asset and authorization boundaries

`AssetReader.Open(context.Context, filename)` returns an `Asset` containing a
seekable, closeable stream and its modification time. These contracts use only
the standard library. The caller closes every non-nil returned stream, including
one returned alongside an error.

`NewDirectoryReader` adapts host-owned files, checks request cancellation and
accepts only single-component filenames for regular files. Bootstrap explicitly
injects it through `NewHandlerWithReader`. `NewHandler(basePath)` remains a
compatible convenience constructor.

The portable service retains the exact public-name allowlist:

| Public name | Host asset |
| --- | --- |
| geo_water | uga_water.geojson |
| geo_districts | uganda_districts.geojson |
| geo_subcounties | uganda_subcounties.geojson |

Unrecognized names return the existing `NOT_FOUND` response without touching
storage. Missing assets and reader failures return `GEOJSON_LOAD_FAILED`
without exposing internal paths. Streams close after successful reads,
conditional requests, ranges and error responses.

Authentication and the existing any-permission gate remain in the host route
registrar: `surveillance:read`, `data_quality:read` or `report_browser:read`.
These checks execute before the reader is invoked. The package extraction must
not replace authoritative host permission resolution.

The response remains raw `application/geo+json`, with
`Cache-Control: public, max-age=86400` and `X-Content-Type-Options: nosniff`.
`http.ServeContent` retains modification-time validation and byte-range behavior.
Map assets and their Docker packaging remain host-owned.

## Phase 2: portable resolution and error handling

`ResolveLayer` owns the immutable, case-sensitive public-name mapping above.
`Service.Open` resolves names before storage access, forwards the request context
to the reader and returns a `Layer` with its canonical filename, modification
time and stream. Successful calls transfer stream ownership to the caller.
Failed or canceled reads close any returned stream and return an empty layer,
including when a reader returns both content and an error.

Unknown names match `ErrUnknownLayer`. Reader failures, absent readers, empty
content and canceled reads match `ErrLoadFailed`. Load errors expose a stable
message while preserving original and close-error causes for `errors.Is/As`.
Thus even an absent file for a recognized layer remains a load failure.

The HTTP handler now delegates to the service and maps only resolution failures
to the existing 404 response. Storage failures retain the existing 500 response,
even when a reader's underlying error also matches `ErrUnknownLayer`.
HTTP envelopes, headers, streaming and permissions stay in the host.
`assets.go`, `service.go` and the portable service tests build and run without
Gin, host files, database drivers or external dependencies.

## Verification and next phases

Tests cover real bundled layers, permission denial before asset access,
allowlist rejection, safe storage errors, stream closure, range/conditional
responses, context cancellation and directory-adapter boundaries. Registry
tests retain enabled/disabled route coverage.

```sh
go test -race ./internal/features/geojson ./internal/api ./internal/bootstrap
make contracts-check
```

The next phases are:

1. Build and verify the standalone module and an external consumer.
2. Publish a version, pin its verified checksums and replace the retained
   host implementation with adapters.
