// Package audit provides self-contained audit reads, PostgreSQL queries, HTTP
// responses, typed caching, integrity manifests and CSV exports.
// Hosts supply verified identity, permissions, export rate enforcement, optional
// cache and caller-owned database pools. Core event recording and production
// schema migrations remain host responsibilities. The package creates no
// workers or storage resources and imports no HTTP framework or SQL driver.
package audit
