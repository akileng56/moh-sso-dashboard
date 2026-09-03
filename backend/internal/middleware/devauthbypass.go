package middleware

// LOCAL DEVELOPMENT ONLY -- DO NOT COMMIT.
//
// Enables running the portal with no Keycloak login at all. Turned on with
// DEV_AUTH_BYPASS=true in backend/app.env (which is gitignored), and is OFF
// unless that variable is explicitly set, so it cannot take effect anywhere
// it has not been deliberately enabled.

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/keycloak"
)

// DevAuthBypassEnabled reports whether the local no-auth mode is switched on.
func DevAuthBypassEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("DEV_AUTH_BYPASS"))) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}

// DevAuthBypassRoutes short-circuits the public session endpoints the frontend
// uses to bootstrap auth, so the portal loads as a signed-in admin with no
// Keycloak round trip. Registered globally, and only when the bypass is on.
func DevAuthBypassRoutes() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := strings.TrimSuffix(c.Request.URL.Path, "/")

		switch {
		case c.Request.Method == http.MethodGet && p == "/api/v1/auth/me":
			response.OK(c, http.StatusOK, gin.H{"user": devBypassUser()})
			c.Abort()
			return

		case p == "/api/v1/auth/login":
			// Skip the Keycloak redirect; land straight back in the portal.
			returnTo := c.Query("returnTo")
			if returnTo == "" {
				returnTo = "/portal/"
			}
			c.Redirect(http.StatusFound, returnTo)
			c.Abort()
			return

		case p == "/api/v1/auth/logout":
			response.OK(c, http.StatusOK, gin.H{"loggedOut": true})
			c.Abort()
			return
		}

		c.Next()
	}
}

// devBypassUser is the synthetic administrator injected on every request while
// the bypass is active.
func devBypassUser() *keycloak.AuthUser {
	return &keycloak.AuthUser{
		ID:            "dev-bypass-user",
		Username:      "dev.admin",
		Email:         "dev.admin@localhost",
		FirstName:     "Dev",
		LastName:      "Admin",
		FullName:      "Dev Admin",
		IsAdmin:       true,
		IsUser:        true,
		RealmRoles:    []string{"admin", "user"},
		ClientRoles:   map[string][]string{},
		Enabled:       true,
		EmailVerified: true,
	}
}

// applyDevBypass populates the same context keys ExtractAuthContext would have
// set, so every downstream handler and guard behaves as if a real admin signed in.
func applyDevBypass(c *gin.Context, resolver authz.PermissionResolver) {
	user := devBypassUser()

	authContext := authz.NewContextWithResolver(
		c.Request.Context(),
		resolver,
		user.ID,
		user.RealmRoles,
		user.ClientRoles,
	)

	c.Set("access_token", "dev-bypass-token")
	c.Set(authz.ContextKey, authContext)
	c.Set("user", user)
	c.Set("user_id", user.ID)
	c.Set("client_roles", authContext.ClientRoles)
	c.Set("realm_roles", authContext.RealmRoles)
	c.Set("permissions", authContext.Permissions)
	c.Set("systems", authContext.AccessibleSystems)
	c.Set("accessible_systems", authContext.AccessibleSystemDetails)
	c.Set("is_admin", true)
	c.Set("is_user", true)

	c.Next()
}
