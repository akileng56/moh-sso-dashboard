// Package geojson owns allowlisted map-layer resolution, embedded map snapshots,
// optional filesystem storage, and a standard-library HTTP module.
//
// NewModule and Register require a trusted platform authorization callback.
// They handle streaming, errors, ranges and conditional reads themselves.
// The lower-level Service API remains available for non-HTTP consumers.
//
// Version v0.2.0 adds the self-contained HTTP registration and embedded assets.
// Hosts supply only shared security/trace capabilities.
package geojson
