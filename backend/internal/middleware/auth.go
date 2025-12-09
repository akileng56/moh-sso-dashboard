package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func ExtractTokenClaims() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		clientRoles := map[string][]string{}

		if resAccess, ok := claims["resource_access"].(map[string]interface{}); ok {
			for clientID, val := range resAccess {
				roleBlock := val.(map[string]interface{})
				if roles, ok := roleBlock["roles"].([]interface{}); ok {
					for _, r := range roles {
						clientRoles[clientID] = append(clientRoles[clientID], r.(string))
					}
				}
			}
		}

		c.Set("client_roles", clientRoles)

		// Is admin?
		if ra, ok := claims["realm_access"].(map[string]interface{}); ok {
			if roles, ok := ra["roles"].([]interface{}); ok {
				for _, r := range roles {
					if r.(string) == "admin" {
						c.Set("is_admin", true)
					}
				}
			}
		}

		c.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ExtractTokenClaims MUST have run before this middleware
		isAdmin, exists := c.Get("is_admin")

		if !exists || isAdmin != true {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
			})
			return
		}

		c.Next()
	}
}
