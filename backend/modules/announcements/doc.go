// Package announcements provides portable announcement lifecycle, audience,
// attachment and delivery orchestration over host-supplied ports.
//
// Hosts authenticate and authorize callers, implement persistence and storage,
// normalize email attachments, and own notification workers and migrations.
// Notification forwarding is best effort and does not make publication an
// atomic or exactly-once delivery operation.
package announcements
