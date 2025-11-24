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

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Base API prefix
	api := r.Group("/api/v1")

	// -----------------------------
	//  PUBLIC ROUTES (NO AUTH)
	// -----------------------------
	auth := api.Group("/auth")
	{
		auth.GET("/login", authHandler.HandleAuthLogin)
		auth.GET("/callback", authHandler.HandleAuthCallback)
		auth.POST("/refresh", authHandler.HandleAuthRefreshToken)
		auth.GET("/logout", authHandler.HandleAuthLogout)
	}

	// -----------------------------
	//  PROTECTED ROUTES (NEED TOKEN)
	// -----------------------------
	protected := api.Group("")
	protected.Use(middleware.ExtractTokenClaims())
	protected.Use(middleware.RequireAuth())

	// Authenticated user info
	protected.GET("/auth/me", authHandler.HandleAuthGetMe)

	// ---- Clients ----
	clients := protected.Group("/clients")
	{
		clients.POST("/", clientHandler.CreateClient)
		clients.GET("/", clientHandler.ListClients)
		clients.GET("/:id", clientHandler.GetClient)
		clients.DELETE("/:id", clientHandler.DeleteClient)
	}

	// ---- Users ----
	users := protected.Group("/users")
	{
		users.POST("/", userHandler.CreateUser)
		users.GET("/", userHandler.ListUsers)
		users.GET("/:id", userHandler.GetUser)
		users.DELETE("/:id", userHandler.DeleteUser)
	}

	return r
}
