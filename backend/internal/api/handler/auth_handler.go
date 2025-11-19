package handler

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
	config      config.Config
}

func NewAuthHandler(authService service.AuthService, config config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		config:      config,
	}
}

func (h *AuthHandler) HandleAuthLogin(c *gin.Context) {
	authURL, err := url.Parse(h.config.KeycloakBaseUrl + "/realms/moh-realm" + "/protocol/openid-connect/auth")
	if err != nil {
		log.Println("Failed to parse Keycloak URL:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	q := authURL.Query()
	q.Set("client_id", h.config.KeycloakClientID)
	q.Set("response_type", "code")
	q.Set("scope", "openid")
	q.Set("redirect_uri", h.config.KeycloakRedirectUri)

	authURL.RawQuery = q.Encode()

	c.Redirect(http.StatusTemporaryRedirect, authURL.String())
}

func (h *AuthHandler) HandleAuthGetMe(c *gin.Context) {
	accessToken, err := c.Cookie("access_token")
	if err != nil || accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	user, err := h.authService.GetMe(accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

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

	// Set cookies
	h.setSecureAccessTokenCookie(c, tokens.AccessToken, tokens.ExpiresIn)
	h.setSecureRefreshTokenCookie(c, tokens.RefreshToken, tokens.RefreshExpiresIn)

	// Redirect to frontend
	c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000/")
}

func (h *AuthHandler) HandleAuthRefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session expired. Please log in."})
		return
	}

	tokens, err := h.authService.GetAccessToken(refreshToken)
	if err != nil {
		log.Printf("Token refresh failed: %v", err)

		// Clear cookie
		h.setSecureRefreshTokenCookie(c, "", -1)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session expired or invalid."})
		return
	}

	// Set new cookies
	h.setSecureAccessTokenCookie(c, tokens.AccessToken, tokens.ExpiresIn)
	h.setSecureRefreshTokenCookie(c, tokens.RefreshToken, tokens.RefreshExpiresIn)

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
		"expires_in":   tokens.ExpiresIn,
	})
}

func (h *AuthHandler) setSecureAccessTokenCookie(c *gin.Context, token string, maxAge int64) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
		Path:     "/",
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) setSecureRefreshTokenCookie(c *gin.Context, token string, maxAge int64) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}
