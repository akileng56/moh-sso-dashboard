// Package geojson resolves the dashboard's fixed map-layer names and reads
// seekable assets through a caller-supplied AssetReader.
//
// Hosts authenticate and authorize calls, supply asset storage, close successful
// streams and own HTTP serving, headers, error envelopes and asset packaging.
// This package contains no map files, storage adapter, HTTP framework, database
// integration, background worker or migration.
package geojson
