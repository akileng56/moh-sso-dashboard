package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/api/handler"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func SetupRouter(
	authHandler *handler.AuthHandler,
	clientHandler *handler.ClientHandler,
	userHandler *handler.UserHandler,
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

	auth := api.Group("/auth")
	{
		auth.GET("/login", authHandler.HandleAuthLogin)
		auth.GET("/callback", authHandler.HandleAuthCallback)
		auth.POST("/refresh", authHandler.HandleAuthRefreshToken)
		auth.GET("/logout", authHandler.HandleAuthLogout)
	}

	protected := api.Group("")
	{
		protected.Use(middleware.ExtractTokenClaims())
		protected.Use(middleware.RequireAuth())
		protected.GET("/auth/me", authHandler.HandleAuthGetMe)
		clients := protected.Group("/clients")
		{
			clients.GET("/", clientHandler.ListClients)
			clients.GET("/:id", clientHandler.GetClient)
			clients.POST("/", clientHandler.CreateClient)
			clients.DELETE("/:id", clientHandler.DeleteClient)
		}
		users := protected.Group("/users")
		{
			users.GET("/", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.POST("/", userHandler.CreateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
