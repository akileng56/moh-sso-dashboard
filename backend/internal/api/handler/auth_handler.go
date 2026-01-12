package handler

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

type AuthHandler struct {
	authService         service.AuthService
	auditService        *service.AuditService
	notificationService service.NotificationsService
	config              *config.Config
}

func NewAuthHandler(
	authService service.AuthService,
	auditService *service.AuditService,
	notificationService service.NotificationsService,
	config *config.Config,
) *AuthHandler {
	return &AuthHandler{
		authService:         authService,
		auditService:        auditService,
		notificationService: notificationService,
		config:              config,
	}
}

// ----------------------------------------------------
// LOGIN (redirect to Keycloak)
// ----------------------------------------------------
func (h *AuthHandler) HandleAuthLogin(c *gin.Context) {

	_ = h.auditService.LoginInitiated(
		c.Request.Context(),
		c.ClientIP(),
		c.Request.UserAgent(),
		h.config.KeycloakClientID,
	)

	authURL, err := url.Parse(
		h.config.KeycloakBaseUrl +
			"/realms/" + h.config.KeycloakRealm +
			"/protocol/openid-connect/auth",
	)
	if err != nil {
		log.Println("failed to parse Keycloak URL:", err)
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

// ----------------------------------------------------
// GET CURRENT USER
// ----------------------------------------------------
func (h *AuthHandler) HandleAuthGetMe(c *gin.Context) {

	accessToken, err := c.Cookie("access_token")
	if err != nil || accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	user, err := h.authService.GetMe(accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	userID := utils.ExtractUserIDFromJWT(accessToken)

	_ = h.auditService.Log(
		c.Request.Context(),
		utils.ToNullUUID(userID),
		"auth.get_me",
		map[string]interface{}{
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
	)

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// ----------------------------------------------------
// OIDC CALLBACK
// ----------------------------------------------------
func (h *AuthHandler) HandleAuthCallback(c *gin.Context) {

	code := c.Query("code")
	if code == "" {

		_ = h.auditService.LoginResult(
			c.Request.Context(),
			uuid.NullUUID{},
			false,
			h.config.KeycloakClientID,
			c.ClientIP(),
			c.Request.UserAgent(),
			"",
			"",
		)

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing authorization code",
		})
		return
	}

	tokens, err := h.authService.ProcessAuthCode(code)
	if err != nil {

		// 🔔 Notification: login failed (best-effort)
		h.notificationService.NotifyLoginFailed(
			c.Request.Context(),
			h.config.KeycloakClientID,
			c.ClientIP(),
			c.Request.UserAgent(),
			err,
		)

		_ = h.auditService.LoginResult(
			c.Request.Context(),
			uuid.NullUUID{},
			false,
			h.config.KeycloakClientID,
			c.ClientIP(),
			c.Request.UserAgent(),
			"",
			"",
		)

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "authentication failed. please try again",
		})
		return
	}

	userID := utils.ExtractUserIDFromJWT(tokens.AccessToken)

	// login success
	_ = h.auditService.LoginResult(
		c.Request.Context(),
		utils.ToNullUUID(userID),
		true,
		h.config.KeycloakClientID,
		c.ClientIP(),
		c.Request.UserAgent(),
		"",
		"",
	)

	h.setSecureAccessTokenCookie(c, tokens.AccessToken, tokens.ExpiresIn)
	h.setSecureRefreshTokenCookie(c, tokens.RefreshToken, tokens.RefreshExpiresIn)

	// Determine admin status
	isAdmin := utils.TokenHasRealmRole(tokens.AccessToken, "admin")

	redirectURL := "http://localhost:3000/dashboard"
	if isAdmin {
		redirectURL = "http://localhost:3000/admin"
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// ----------------------------------------------------
// REFRESH TOKEN
// ----------------------------------------------------
func (h *AuthHandler) HandleAuthRefreshToken(c *gin.Context) {

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {

		_ = h.auditService.TokenRefresh(
			c.Request.Context(),
			uuid.NullUUID{},
			false,
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "session expired",
		})
		return
	}

	tokens, err := h.authService.GetAccessToken(refreshToken)
	if err != nil {

		// 🔔 Notification: token refresh failed (best-effort)
		h.notificationService.NotifyTokenRefreshFailed(
			c.Request.Context(),
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		_ = h.auditService.TokenRefresh(
			c.Request.Context(),
			uuid.NullUUID{},
			false,
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		h.setSecureRefreshTokenCookie(c, "", -1)

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid refresh token",
		})
		return
	}

	userID := utils.ExtractUserIDFromJWT(tokens.AccessToken)

	_ = h.auditService.TokenRefresh(
		c.Request.Context(),
		utils.ToNullUUID(userID),
		true,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	h.setSecureAccessTokenCookie(c, tokens.AccessToken, tokens.ExpiresIn)
	h.setSecureRefreshTokenCookie(c, tokens.RefreshToken, tokens.RefreshExpiresIn)

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
		"expires_in":   tokens.ExpiresIn,
	})
}

// ----------------------------------------------------
// LOGOUT
// ----------------------------------------------------
func (h *AuthHandler) HandleAuthLogout(c *gin.Context) {

	userID := utils.ToNullUUID(c.GetString("user_id"))

	_ = h.auditService.Logout(
		c.Request.Context(),
		userID,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	refreshToken, _ := c.Cookie("refresh_token")
	if refreshToken != "" {
		_ = h.authService.LogOut(refreshToken)
	}

	clearCookie := func(name string, httpOnly bool) {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: httpOnly,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
	}

	clearCookie("access_token", false)
	clearCookie("refresh_token", true)

	c.Redirect(
		http.StatusTemporaryRedirect,
		"http://localhost:9000/api/v1/auth/login",
	)
}

// ----------------------------------------------------
// COOKIE HELPERS
// ----------------------------------------------------
func (h *AuthHandler) setSecureAccessTokenCookie(
	c *gin.Context,
	token string,
	maxAge int64,
) {
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

func (h *AuthHandler) setSecureRefreshTokenCookie(
	c *gin.Context,
	token string,
	maxAge int64,
) {
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
