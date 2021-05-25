package cli

import (
	"context"
	"github.com/alecthomas/kingpin"
	"github.com/bsc-bridge-svc/internal/config"
	"github.com/bsc-bridge-svc/internal/data/migrate"
	"github.com/bsc-bridge-svc/internal/services/server"
	"os"
)

func Run(args []string) bool {
	cfg := config.New(os.Getenv("CONFIG"))
	ctx := context.Background()
	log := cfg.Logger()

	defer func() {
		if rvr := recover(); rvr != nil {
			log.Error("app panicked\n", rvr)
		}
	}()

	app := kingpin.New("bsc-bridge-svc", "")

	runCmd := app.Command("run", "run command")
	migrateDBCmd := app.Command("migrate", "migrate command")
	migrateDBUpCmd := migrateDBCmd.Command(migrate.Up, "migrate db up")
	migrateDBDownCmd := migrateDBCmd.Command(migrate.Down, "migrate db down")

	cmd, err := app.Parse(args[1:])
	if err != nil {
		log.WithError(err).Error("failed to parse arguments")
		return false
	}

	switch cmd {
	case runCmd.FullCommand():
		if err := server.New(cfg, ctx).Run(); err != nil {
			log.WithError(err).Error("failed to run bridge service")
			return false
		}
	case migrateDBUpCmd.FullCommand():
		_, err = migrate.MigrateUp(cfg)
	case migrateDBDownCmd.FullCommand():
		_, err = migrate.MigrateDown(cfg)
	default:
		log.WithField("command", cmd).Error("Unknown command")
		return false
	}

	return true
}
