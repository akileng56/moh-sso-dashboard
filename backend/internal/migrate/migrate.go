package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Required for file source
	"github.com/rs/zerolog/log"
)

func MigrateDB(db *sql.DB, migrateDir string) error {
	start := time.Now()
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrateDir,
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Info().Msgf("Database migrated successfully in %s", time.Since(start))
	return nil
}

func DropDB(db *sql.DB, migrateDir string) error {
	start := time.Now()
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrateDir,
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("drop migration failed: %w", err)
	}

	log.Info().Msgf("Database dropped successfully in %s", time.Since(start))
	return nil
}

func PingDB(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}
func WaitForDB(ctx context.Context, db *sql.DB) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for database: %w", ctx.Err())
		case <-ticker.C:
			if err := PingDB(ctx, db); err == nil {
				log.Info().Msg("Database is up and running")
				return nil
			} else {
				log.Warn().Err(err).Msg("Waiting for database to be ready...")
			}
		}
	}
}
