package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/service"
)

// AuthHandler handles HTTP requests related to authentication
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new handler instance
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// HandleAuthCallback is the endpoint that receives the Authorization Code from Keycloak
func (h *AuthHandler) HandleAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}
	tokens, err := h.authService.ProcessAuthCode(code)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed. Please try again."})
		return
	}

	// Use Gin's SetCookie method for a cleaner API
	c.SetCookie(
		"refresh_token",              // name
		tokens.RefreshToken,          // value
		int(tokens.RefreshExpiresIn), // maxAge (Gin uses seconds)
		"/",                          // path
		"localhost",                  // domain (use the hostname your frontend is served from)
		false,                        // secure (true for HTTPS/localhost)
		true,                         // httpOnly
	)

	// Gin's SetCookie doesn't expose the SameSite option easily.
	// To ensure SameSite is set explicitly, we'll revert to the standard net/http way
	// using c.Writer and explicitly set the cookie object.

	// 🚨 IMPORTANT: Override to ensure explicit SameSite=Lax
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Expires:  time.Now().Add(time.Duration(tokens.RefreshExpiresIn) * time.Second),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	// 4. Redirect the user back to the main dashboard/frontend URL
	c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000/dashboard")
}
