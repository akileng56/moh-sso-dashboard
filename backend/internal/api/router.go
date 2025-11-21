package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/api/handler"
)

func SetupRouter(
	authHandler *handler.AuthHandler,
	clientHandler *handler.ClientHandler,
	userHandler *handler.UserHandler,
) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api/v1")
	{
		// ---- auth ----
		auth := api.Group("/auth")
		{
			auth.GET("/callback", authHandler.HandleAuthCallback)
			auth.POST("/refresh", authHandler.HandleAuthRefreshToken)
			auth.GET("/login", authHandler.HandleAuthLogin)
			auth.GET("/me", authHandler.HandleAuthGetMe)
			auth.GET("/logout", authHandler.HandleAuthLogout)
		}

		// --- Clients ---
		clients := api.Group("/clients")
		{
			clients.POST("/", clientHandler.CreateClient)
			clients.GET("/", clientHandler.ListClients)
			clients.GET("/:id", clientHandler.GetClient)
			clients.DELETE("/:id", clientHandler.DeleteClient)
		}

		// --- Users ---
		users := api.Group("/users")
		{
			users.POST("/", userHandler.CreateUser)
			users.GET("/", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	return r
}
