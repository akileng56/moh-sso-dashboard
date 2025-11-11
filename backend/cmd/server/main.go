package main

import (
	"database/sql"
	"log"

	router "github.com/moh-sso-dashboard/internal/api"
	"github.com/moh-sso-dashboard/internal/api/handler"
	config "github.com/moh-sso-dashboard/internal/config"
	logger "github.com/moh-sso-dashboard/internal/log"
	clientRepo "github.com/moh-sso-dashboard/internal/repository/client"
	userRepo "github.com/moh-sso-dashboard/internal/repository/user"
	"github.com/moh-sso-dashboard/internal/service"

	"github.com/rs/zerolog"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}
	appLogger := logger.NewLogger()
	appLogger.SetLevel(zerolog.InfoLevel)

	appLogger.Info("Starting server in environment: %s", cfg.Environment)

	conn, err := sql.Open(cfg.DbDriver, cfg.DbSource())
	if err != nil {
		appLogger.Fatal("Cannot open database connection")
	}
	defer conn.Close() 

	err = conn.Ping()
	if err != nil {
		appLogger.Fatal("Cannot connect to database")
	}

	appLogger.Info("Successfully connected to database")

	clientRepo := clientRepo.NewClientRepository(conn) 
	userRepo := userRepo.NewUserRepository(conn)

	// Services
	clientService := service.NewClientService(clientRepo)
	userService := service.NewUserService(userRepo)

	// Handlers
	clientHandler := handler.NewClientHandler(clientService)
	userHandler := handler.NewUserHandler(userService)

	r := router.SetupRouter(clientHandler, userHandler)

	appLogger.Info("Server listening on port :%s", cfg.DbPort) 
	
	// Run the Gin server
	if err := r.Run(":" + cfg.DbPort); err != nil {
		appLogger.Fatal("Gin server failed to run")
	}
}