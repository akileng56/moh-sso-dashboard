// Package reportscheduler provides report scheduling, Health BI generation,
// delivery, retry/recovery and a cancellable worker without host dependencies.
//
// Callers own authentication, permissions, ownership/scope inputs and HTTP
// adapters. Repository persists scheduler state; the supplied PostgreSQL adapter
// reads the host-owned scheduler schema, user directory, groups and DWH hierarchy.
// This package neither applies migrations nor starts workers implicitly.
package reportscheduler
