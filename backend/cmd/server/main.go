package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	router "github.com/moh-sso-dashboard/internal/api"
	"github.com/moh-sso-dashboard/internal/api/handler"
	config "github.com/moh-sso-dashboard/internal/config"
	kcClientPkg "github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
	db "github.com/moh-sso-dashboard/internal/migrate"
	authRepo "github.com/moh-sso-dashboard/internal/repository/auth"
	clientRepo "github.com/moh-sso-dashboard/internal/repository/client"
	metricsRepo "github.com/moh-sso-dashboard/internal/repository/metrics"
	userRepo "github.com/moh-sso-dashboard/internal/repository/user"
	"github.com/moh-sso-dashboard/internal/service"

	redis "github.com/moh-sso-dashboard/internal/cache"
	store "github.com/moh-sso-dashboard/internal/db/sqlc"

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

	if err := db.MigrateDB(conn, "file://internal/db/migrations"); err != nil {
		appLogger.Fatal("Cannot migrate db: %v", err)
	}

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

	// store
	store := store.NewStore(conn)

	// redis
	rdb := redis.NewRedisClient(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword)

	// --- Repository Layer Initialization ---
	authRepo := authRepo.NewAuthRepository(keycloakClient, cfg)
	clientRepo := clientRepo.NewClientRepository(keycloakClient, cfg, store, *appLogger)
	userRepo := userRepo.NewUserRepository(conn)
	metrics := metricsRepo.NewMetricsRepository(cfg, store, *appLogger)

	// --- Service Layer Initialization ---
	clientService := service.NewClientService(clientRepo)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(authRepo, rdb)
	metricsService := service.NewMetricsService(metrics)

	// --- Handler Layer Initialization ---
	clientHandler := handler.NewClientHandler(clientService)
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService, cfg)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	// heallthHandler := handler.NewHealthHandler()

	// --- Router and Server Start ---
	r := router.SetupRouter(authHandler, clientHandler, userHandler, metricsHandler)

	appLogger.Info("Server listening securely on port :%s", cfg.ServerPort)

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		appLogger.Fatal("Gin server failed to run with TLS: %v", err)
	}
}
