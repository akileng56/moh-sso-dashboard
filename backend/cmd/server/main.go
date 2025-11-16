package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // Database driver import
	router "github.com/moh-sso-dashboard/internal/api"
	"github.com/moh-sso-dashboard/internal/api/handler"
	config "github.com/moh-sso-dashboard/internal/config"
	kcClientPkg "github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
	authRepo "github.com/moh-sso-dashboard/internal/repository/auth"
	clientRepo "github.com/moh-sso-dashboard/internal/repository/client"
	userRepo "github.com/moh-sso-dashboard/internal/repository/user"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/rs/zerolog"
)

func main() {
	// --- Configuration and Logger Setup ---
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	appLogger := logger.NewLogger()
	appLogger.SetLevel(zerolog.InfoLevel)
	appLogger.Info("Starting server in environment: %s", cfg.Environment)

	// --- Database Setup ---
	conn, err := sql.Open(cfg.DbDriver, cfg.DbSource())
	if err != nil {
		appLogger.Fatal("Cannot open database connection: %v", err)
	}
	if conn == nil {
		appLogger.Fatal("Database connection is nil — check your configuration.")
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		appLogger.Fatal("Cannot connect to database: %v", err)
	}
	appLogger.Info("Successfully connected to database")

	keycloakClient := kcClientPkg.NewClient(
		cfg.KeycloakBaseUrl,
		cfg.KeycloakRealm,
		cfg.KeycloakClientID,
		cfg.KeycloakClientSecret,
	)

	if err := keycloakClient.Authenticate(); err != nil {
		appLogger.Fatal("Failed to authenticate Keycloak service account: %v", err)
	}
	appLogger.Info("Successfully authenticated Keycloak service account.")

	// --- Repository Layer Initialization ---
	authRepo := authRepo.NewAuthRepository(keycloakClient, cfg)
	clientRepo := clientRepo.NewClientRepository(conn)
	userRepo := userRepo.NewUserRepository(conn)

	// --- Service Layer Initialization ---
	clientService := service.NewClientService(clientRepo)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(authRepo)

	// --- Handler Layer Initialization ---
	clientHandler := handler.NewClientHandler(clientService)
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)

	// --- Router and Server Start ---
	r := router.SetupRouter(authHandler, clientHandler, userHandler)

	// 🚨 FIX: Change log message to reflect HTTPS
	appLogger.Info("Server listening securely on port :%s", cfg.ServerPort)

	// 🚨 FIX: Use r.RunTLS to enable HTTPS on the Gin router
	if err := r.RunTLS(":"+cfg.ServerPort, cfg.TlsCert, cfg.TlsKey); err != nil {
		appLogger.Fatal("Gin server failed to run with TLS: %v", err)
	}
}
