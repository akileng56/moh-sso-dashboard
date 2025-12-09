package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/api/handler"
	"github.com/moh-sso-dashboard/internal/middleware"

	service "github.com/moh-sso-dashboard/internal/service"
)

func SetupRouter(
	authHandler *handler.AuthHandler,
	clientHandler *handler.ClientHandler,
	userHandler *handler.UserHandler,
	metricsHandler *handler.MetricsHandler,
	auditSvc *service.AuditService,
) *gin.Engine {

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api/v1")
	api.Use(middleware.AuditMiddleware(auditSvc))

	// ----------------------
	// Public Auth Endpoints
	// ----------------------
	auth := api.Group("/auth")
	{
		auth.GET("/login", authHandler.HandleAuthLogin)
		auth.GET("/callback", authHandler.HandleAuthCallback)
		auth.POST("/refresh", authHandler.HandleAuthRefreshToken)
		auth.GET("/logout", authHandler.HandleAuthLogout)
	}

	// ----------------------
	// Protected API
	// ----------------------
	protected := api.Group("")
	protected.Use(middleware.RequireAuth())
	protected.Use(middleware.ExtractTokenClaims())
	{
		protected.GET("/auth/me", authHandler.HandleAuthGetMe)

		// Clients
		clients := protected.Group("/clients")
		{
			clients.GET("/", clientHandler.ListClients)
			clients.GET("/:id", clientHandler.GetClient)
			clients.POST("/", clientHandler.CreateClient)
			clients.DELETE("/:id", clientHandler.DeleteClient)
		}

		// Users
		users := protected.Group("/users")
		{
			users.GET("/", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.POST("/", userHandler.CreateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// ----------------------
		// Admin Section
		// ----------------------
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireAdmin())
		{
			// User Admin Management
			admin.GET("/users", userHandler.ListUsers)
			admin.GET("/users/:id", userHandler.GetUser)
			admin.POST("/users", userHandler.CreateUser)
			admin.DELETE("/users/:id", userHandler.DeleteUser)

			// =====================
			// METRICS SECTION
			// =====================
			metrics := admin.Group("/metrics")
			{
				// OVERVIEW
				metrics.GET("/overview", metricsHandler.Overview)

				// SYSTEM
				metrics.GET("/system/count-users", metricsHandler.CountUsers)
				metrics.GET("/system/count-disabled-users", metricsHandler.CountDisabledUsers)
				metrics.GET("/system/active-today", metricsHandler.ActiveUsersToday)
				metrics.GET("/system/active-this-week", metricsHandler.ActiveUsersThisWeek)
				metrics.GET("/system/login-trend", metricsHandler.LoginTrend)
				metrics.GET("/system/login-trend-range", metricsHandler.LoginTrendByDay)

				// SECURITY
				metrics.GET("/security/failed-logins", metricsHandler.CountFailedLogins)
				metrics.GET("/security/failed-logins-range", metricsHandler.CountFailedLoginsInRange)
				metrics.GET("/security/suspicious-logins", metricsHandler.SuspiciousLogins)

				// CLIENTS
				metrics.GET("/clients/count", metricsHandler.CountClients)
				metrics.GET("/clients/most-accessed", metricsHandler.MostAccessedClients)
				metrics.GET("/clients/login-count", metricsHandler.LoginCountForClient)
				metrics.GET("/clients/active-today", metricsHandler.ActiveUsersPerClientToday)

				// USERS
				metrics.GET("/users/new-range", metricsHandler.NewUsersInRange)
				metrics.GET("/users/new-trend", metricsHandler.NewUsersTrend)
				metrics.GET("/users/never-logged-in", metricsHandler.NeverLoggedInUsers)
				metrics.GET("/users/last-login/:userID", metricsHandler.LastLoginForUser)
				metrics.GET("/users/client-usage/:userID", metricsHandler.UserClientUsage)
			}
		}
	}

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
