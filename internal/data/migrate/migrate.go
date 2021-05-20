package migrate

import (
	"github.com/bsc-bridge-svc/internal/assets"
	"github.com/bsc-bridge-svc/internal/config"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
)

const (
	Up   = "up"
	Down = "down"
)

// package for database migration
// stores migrations to binaries
var migrations = &migrate.PackrMigrationSource{
	Box: assets.Migrations,
}

// MigrateUp Migrates database up
func MigrateUp(cfg config.Config) (int, error) {
	applied, err := migrate.Exec(cfg.DB(), "postgres", migrations, migrate.Up)
	if err != nil {
		return 0, errors.Wrap(err, "failed to apply migrations")
	}

	cfg.Logger().WithField("applied", applied).Info("Migrations applied")

	return applied, nil
}

// MigrateDown Migrates database down
func MigrateDown(cfg config.Config) (int, error) {
	applied, err := migrate.Exec(cfg.DB(), "postgres", migrations, migrate.Down)
	if err != nil {
		return 0, errors.Wrap(err, "failed to apply migrations")
	}

	cfg.Logger().WithField("applied", applied).Info("Migrations applied")

	return applied, nil
}
