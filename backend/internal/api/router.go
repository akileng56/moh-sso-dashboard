package router

import (
	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/api/handler"
)

// SetupRouter initializes and configures the Gin router with all handlers.
func SetupRouter(
	clientHandler *handler.ClientHandler,
	userHandler *handler.UserHandler,
) *gin.Engine {
	r := gin.Default()

	// Use the standard CORS setup if needed for cross-origin requests
	// r.Use(cors.New(cors.Config{...}))

	// Define the base API group
	api := r.Group("/api/v1")
	{
		// --- Client Management Routes ---
		clients := api.Group("/clients")
		{
			// FIX: Renamed CreateClientHandler to CreateClient to match the Gin method signature
			clients.POST("/", clientHandler.CreateClient)
			clients.GET("/", clientHandler.ListClients)
			clients.GET("/:id", clientHandler.GetClient)
			clients.DELETE("/:id", clientHandler.DeleteClient)
		}

		// --- User Management Routes ---
		users := api.Group("/users")
		{
			// All user methods already match the standardized pattern
			users.POST("/", userHandler.CreateUser)
			users.GET("/", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	return r
}