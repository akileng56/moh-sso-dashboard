package router

import (
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
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowCredentials = true
	config.AddAllowHeaders("Authorization")

	r.Use(cors.New(config))
	// Define the base API group
	api := r.Group("/api/v1")
	{

		// ---- auth management ------
		auth := api.Group("/auth")
		{
			auth.GET("/callback", authHandler.HandleAuthCallback)
			auth.POST("/refresh", authHandler.HandleAuthRefreshToken)
			auth.GET("/login", authHandler.HandleAuthLogin)
			auth.GET("/me", authHandler.HandleAuthGetMe)
			auth.POST("/logout")
		}

		// --- Client Management Routes ---
		clients := api.Group("/clients")
		{
			clients.POST("/", clientHandler.CreateClient)
			clients.GET("/", clientHandler.ListClients)
			clients.GET("/:id", clientHandler.GetClient)
			clients.DELETE("/:id", clientHandler.DeleteClient)
		}

		// --- User Management Routes ---
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
