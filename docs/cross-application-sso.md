# Cross-Application SSO

The Integrated Health Portal uses Keycloak for identity and a backend-managed `sso_session` cookie for its application session. Opening `/portal/` directly is intentionally public and does not create that backend session.

## Launch Contract

Applications that launch the portal and need an authenticated destination must use the registry's `authenticatedLaunchUrl`:

```text
/api/v1/auth/launch?returnTo=/portal/apps/news
```

The ordinary `launchUrl` remains `/portal`. The launcher resolves the authenticated backend path against its runtime API base. In production, an API base of `https://dashboards.health.go.ug/ssobackend` therefore produces:

```text
https://dashboards.health.go.ug/ssobackend/api/v1/auth/launch?returnTo=%2Fportal%2Fapps%2Fnews
```

The backend redirects to Keycloak. An existing Keycloak SSO session is reused, the fixed callback creates the portal `sso_session`, and the browser returns to the validated portal route. Tokens, authorization codes, and session identifiers are never added to launcher URLs.

## Registry Metadata

The `dashboard-web` system should define:

```yaml
launchUrl: /portal
authenticatedLaunchUrl: /api/v1/auth/launch?returnTo=%2Fportal%2Fapps%2Fnews
```

The equivalent Keycloak client attributes are `ui.launchUrl` and `ui.authenticatedLaunchUrl`. Portal administrators can edit both values through Systems or RBAC Management. Report Browser and other external launchers should consume this metadata instead of hardcoding `/portal/`.

## Redirect Security

`returnTo` accepts `/portal` paths and optional query strings or fragments. Absolute URLs require an exact origin in `AUTH_RETURN_URL_ALLOWED_ORIGINS`. The backend rejects foreign origins, credentials, unsupported schemes, scheme-relative URLs, control characters, encoded separators, double encoding, and dot-segment traversal. The target is bound to OAuth state in a short-lived HttpOnly, SameSite=Lax cookie and is cleared after success or terminal failure.

HTTP `Referer` is never used as a redirect target.

## Environment Configuration

Development:

```env
FRONTEND_BASE_URL=http://localhost:3000/portal
AUTH_RETURN_URL_ALLOWED_ORIGINS=http://localhost:3000
KEYCLOAK_REDIRECT_URI=http://localhost:9000/api/v1/auth/callback
```

Production:

```env
FRONTEND_BASE_URL=https://dashboards.health.go.ug/portal
AUTH_RETURN_URL_ALLOWED_ORIGINS=https://dashboards.health.go.ug
KEYCLOAK_REDIRECT_URI=https://dashboards.health.go.ug/ssobackend/api/v1/auth/callback
API_BASE_URL=/ssobackend
```

The portal and the launching application must use the same Keycloak realm and browser-visible Keycloak origin. Configure only the exact callback URI. Avoid broad redirect wildcards.

## Verification

1. Sign into Report Browser through Keycloak.
2. Launch Integrated Health Portal using `authenticatedLaunchUrl`.
3. Confirm Keycloak does not ask for credentials again.
4. Confirm the browser arrives at `/portal/apps/news`.
5. Confirm `/ssobackend/api/v1/auth/me` returns the authenticated user.
6. Confirm portal menus reflect that user's effective RBAC permissions.

If the public News page appears unauthenticated, inspect the launcher URL first. A direct `/portal/` link bypasses portal session creation by design.
