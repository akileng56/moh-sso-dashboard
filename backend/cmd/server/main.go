package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"

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
	"github.com/moh-sso-dashboard/internal/repository/notifications"
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

	// ---------------------------------------------------------------------
	// Database setup
	// ---------------------------------------------------------------------
	conn, err := sql.Open(cfg.DbDriver, cfg.DbSource())
	if err != nil {
		appLogger.Fatal("Cannot open database connection: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		appLogger.Fatal("Cannot connect to database: %v", err)
	}
	appLogger.Info("Successfully connected to database")

	if err := db.MigrateDB(conn, "file://internal/db/migrations"); err != nil {
		appLogger.Fatal("Cannot migrate db: %v", err)
	}

	// ---------------------------------------------------------------------
	// Keycloak
	// ---------------------------------------------------------------------
	keycloakClient := kcClientPkg.NewClient(
		cfg.KeycloakBaseUrl,
		cfg.KeycloakRealm,
		cfg.KeycloakClientID,
		cfg.KeycloakClientSecret,
	)

	if err := keycloakClient.Authenticate(); err != nil {
		appLogger.Fatal("Failed to authenticate Keycloak service account: %v", err)
	}
	appLogger.Info("Successfully authenticated Keycloak service account")

	// ---------------------------------------------------------------------
	// Infrastructure
	// ---------------------------------------------------------------------
	store := store.NewStore(conn)
	rdb := redis.NewRedisClient(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword)

	// ---------------------------------------------------------------------
	// Repositories
	// ---------------------------------------------------------------------
	authRepository := authRepo.NewAuthRepository(keycloakClient, cfg)
	clientRepository := clientRepo.NewClientRepository(keycloakClient, cfg, store, *appLogger)
	userRepository := userRepo.NewUserRepository(keycloakClient, cfg, store, *appLogger)
	metricsRepository := metricsRepo.NewMetricsRepository(cfg, store, *appLogger)
	notificationsRepository := notifications.NewNotificationsRepository(store, *appLogger)

	// ---------------------------------------------------------------------
	// Services
	// ---------------------------------------------------------------------
	clientService := service.NewClientService(clientRepository)
	userService := service.NewUserService(userRepository)
	authService := service.NewAuthService(authRepository, rdb)
	metricsService := service.NewMetricsService(metricsRepository)
	auditService := service.NewAuditService(store)
	importService := service.NewImportService(store, keycloakClient)
	notificationsService := service.NewNotificationsService(notificationsRepository)

	// ---------------------------------------------------------------------
	// Handlers
	// ---------------------------------------------------------------------
	clientHandler := handler.NewClientHandler(clientService, auditService)
	userHandler := handler.NewUserHandler(userService, auditService)
	authHandler := handler.NewAuthHandler(authService, auditService, cfg)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	importHandler := handler.NewImportHandler(importService, cfg)
	auditHandler := handler.NewAuditHandler(store)
	notificationsHandler := handler.NewNotificationsHandler(notificationsService)

	// ---------------------------------------------------------------------
	// Router
	// ---------------------------------------------------------------------
	r := router.SetupRouter(
		importHandler,
		authHandler,
		clientHandler,
		userHandler,
		metricsHandler,
		auditService,
		auditHandler,
		notificationsHandler,
	)

	// ---------------------------------------------------------------------
	// 🚀 Server start (FORCED IPv4 — FIXES ECONNREFUSED)
	// ---------------------------------------------------------------------
	addr := ":" + cfg.ServerPort

	ln, err := net.Listen("tcp4", addr) // 🔥 FORCE IPv4
	if err != nil {
		appLogger.Fatal("Failed to bind IPv4 listener: %v", err)
	}

	appLogger.Info("Gin server listening on IPv4 %s", addr)

	server := &http.Server{
		Handler: r,
	}

	if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
		appLogger.Fatal("Gin server failed: %v", err)
	}
}
