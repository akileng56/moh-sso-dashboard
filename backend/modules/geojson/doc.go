// Package geojson owns allowlisted map-layer resolution, embedded map snapshots,
// optional filesystem storage, and a standard-library HTTP module.
//
// NewModule and Register require a trusted platform authorization callback.
// They handle streaming, errors, ranges and conditional reads themselves.
// The lower-level Service API remains available for non-HTTP consumers.
//
// The self-contained API is an unreleased candidate. The dashboard remains
// pinned to v0.1.0 until the release-and-integration phase.
package geojson
