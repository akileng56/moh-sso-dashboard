// Package announcements provides portable announcement lifecycle, audience,
// attachment and delivery orchestration, plus PostgreSQL persistence.
//
// Hosts authenticate and authorize callers, supply shared identity, storage and
// delivery capabilities, and own infrastructure pools and migrations.
// Notification forwarding is best effort and does not make publication an
// atomic or exactly-once delivery operation.
package announcements
