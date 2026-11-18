// Package audit provides portable audit queries, optional typed read caching,
// export integrity manifests and CSV formatting through host-supplied ports.
//
// Hosts own authentication, permissions, query validation, SQL/cache adapters,
// event recording and schema migrations. The package creates no workers or
// storage resources and has no HTTP or database-driver dependency.
package audit
