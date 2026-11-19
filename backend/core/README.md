# Portable platform core

This host-integrated package owns the standard-library-only inventory,
dependency validation and lifecycle contracts used by compiled modules.
See [the platform core boundaries](../../docs/platform-core.md) for required
capabilities, host adapters and the remaining minimal-boot work.

The existing `backend/module` Gin adapter delegates to this runtime.
Core capabilities are composed by the host and are not optional module IDs.

Run `python3 infra/scripts/test-platform-core.py` from the repository root
to verify the whole core outside the host tree with an empty dependency cache
and workspaces/downloads disabled. The check builds, vets and runs uncached
race tests. This package is not independently published or version-pinned yet.


Phase 2 adds verified-claim identity, generic host session-store and
authorization-resolver/access contracts. Permission/role/system predicates are
shared without relocating application policy catalogues. Host authentication,
cookie/token parsing, Redis payload types, provider selection and static
defaults remain outside this package. `ResolveAuthorization` discards partial
grants on provider error; it never authenticates credentials or invents grants.
