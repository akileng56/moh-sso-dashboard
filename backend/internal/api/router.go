package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/api/handler"
	"github.com/moh-sso-dashboard/internal/keycloak"
	"github.com/moh-sso-dashboard/internal/middleware"
	service "github.com/moh-sso-dashboard/internal/service"
)

func SetupRouter(
	kcClient *keycloak.Client, // ✅ ADD THIS
	importHandler *handler.ImportHandler,
	authHandler *handler.AuthHandler,
	clientHandler *handler.ClientHandler,
	userHandler *handler.UserHandler,
	metricsHandler *handler.MetricsHandler,
	auditSvc *service.AuditService,
	auditHandler *handler.AuditHandler,
	notificationsHandler *handler.NotificationsHandler,
) *gin.Engine {

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// --------------------------------------------------
	// CORS
	// --------------------------------------------------
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// --------------------------------------------------
	// API
	// --------------------------------------------------
	api := r.Group("/api/v1")

	// --------------------------------------------------
	// Auth (PUBLIC)
	// --------------------------------------------------
	auth := api.Group("/auth")
	{
		auth.GET("/login", authHandler.HandleAuthLogin)
		auth.GET("/callback", authHandler.HandleAuthCallback)
		auth.POST("/refresh", authHandler.HandleAuthRefreshToken)
		auth.GET("/logout", authHandler.HandleAuthLogout)
	}

	// --------------------------------------------------
	// Protected (AUTH REQUIRED, CACHE-BACKED)
	// --------------------------------------------------
	protected := api.Group("")
	protected.Use(middleware.ExtractAuthContext(kcClient)) // ✅ NEW
	protected.Use(middleware.RequireAuth())
	protected.Use(middleware.AuditMiddleware(auditSvc))
	{
		protected.GET("/auth/me", authHandler.HandleAuthGetMe)

		// ------------------
		// Clients
		// ------------------
		clients := protected.Group("/clients")
		{
			clients.GET("", clientHandler.ListClients)
			clients.GET("/:id", clientHandler.GetClient)
			clients.POST("", clientHandler.CreateClient)
			clients.DELETE("/:id", clientHandler.DeleteClient)

			clients.GET("/:id/roles", clientHandler.ListClientRoles)
		}

		// ------------------
		// Users
		// ------------------
		users := protected.Group("/users")
		{
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.POST("", userHandler.CreateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// --------------------------------------------------
		// Admin (ADMIN ONLY)
		// --------------------------------------------------
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireAdmin())
		{
			admin.GET("/users", userHandler.ListUsers)
			admin.GET("/users/:id", userHandler.GetUser)
			admin.POST("/users", userHandler.CreateUser)
			admin.DELETE("/users/:id", userHandler.DeleteUser)

			admin.POST("/users/import/preview", importHandler.Preview)
			admin.POST("/users/import/execute", importHandler.Execute)
			admin.GET("/users/import/:jobId", importHandler.GetJob)
			admin.GET("/users/import/:jobId/errors.csv", importHandler.DownloadErrorsCSV)
			admin.GET("/users/import/template.csv", importHandler.DownloadTemplateCSV)

			// -------- Client roles --------
			admin.POST("/clients/:id/roles", clientHandler.CreateClientRole)
			admin.DELETE("/clients/:id/roles/:role", clientHandler.DeleteClientRole)

			// -------- User ↔ Client roles --------
			admin.POST("/users/:id/clients/:clientId/roles", clientHandler.AssignClientRoleToUser)
			admin.DELETE("/users/:id/clients/:clientId/roles/:role", clientHandler.RemoveClientRoleFromUser)

			// -------- Metrics --------
			metrics := admin.Group("/metrics")
			{
				metrics.GET("/overview", metricsHandler.Overview)

				metrics.GET("/system/count-users", metricsHandler.CountUsers)
				metrics.GET("/system/count-disabled-users", metricsHandler.CountDisabledUsers)
				metrics.GET("/system/active-today", metricsHandler.ActiveUsersToday)
				metrics.GET("/system/active-this-week", metricsHandler.ActiveUsersThisWeek)
				metrics.GET("/system/login-trend", metricsHandler.LoginTrend)
				metrics.GET("/system/login-trend-range", metricsHandler.LoginTrendByDay)

				metrics.GET("/security/failed-logins", metricsHandler.CountFailedLogins)
				metrics.GET("/security/failed-logins-range", metricsHandler.CountFailedLoginsInRange)
				metrics.GET("/security/suspicious-logins", metricsHandler.SuspiciousLogins)

				metrics.GET("/clients/count", metricsHandler.CountClients)
				metrics.GET("/clients/most-accessed", metricsHandler.MostAccessedClients)
				metrics.GET("/clients/login-count", metricsHandler.LoginCountForClient)
				metrics.GET("/clients/active-today", metricsHandler.ActiveUsersPerClientToday)

				metrics.GET("/users/new-range", metricsHandler.NewUsersInRange)
				metrics.GET("/users/new-trend", metricsHandler.NewUsersTrend)
				metrics.GET("/users/never-logged-in", metricsHandler.NeverLoggedInUsers)
				metrics.GET("/users/last-login/:userID", metricsHandler.LastLoginForUser)
				metrics.GET("/users/client-usage/:userID", metricsHandler.UserClientUsage)
			}

			// -------- Audit Logs --------
			audit := admin.Group("/audit-logs")
			{
				audit.GET("", auditHandler.ListAuditLogs)
				audit.GET("/actions", auditHandler.ListAuditActions)
				audit.GET("/:id", auditHandler.GetAuditLog)

				audit.GET("/metrics/overview", auditHandler.AuditMetricsOverview)
				audit.GET("/metrics/failed-logins-by-day", auditHandler.FailedLoginsByDay)
				audit.GET("/metrics/top-failure-ips", auditHandler.TopFailureIPs)

				audit.GET("/export", auditHandler.ExportAuditLogs)
			}

			// -------- Notifications ----------
			notifications := admin.Group("/notifications")
			{
				notifications.POST("", notificationsHandler.Notify)
				notifications.GET("", notificationsHandler.ListNotifications)
				notifications.GET("/:id", notificationsHandler.GetNotificationByID)
				notifications.PATCH("/:id/read", notificationsHandler.MarkNotificationAsRead)
				notifications.DELETE("/:id", notificationsHandler.DeleteNotification)

				notifications.GET("/count", notificationsHandler.CountNotifications)
				notifications.GET("/count/unread", notificationsHandler.CountUnreadNotificationsCount)

				notifications.DELETE("/cleanup", notificationsHandler.DeleteOldNotifications)
			}
		}
	}

	// --------------------------------------------------
	// Health
	// --------------------------------------------------
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
