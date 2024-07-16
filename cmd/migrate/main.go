package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-kit/log"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"ws-dummy-go/internal/app"
)

func main() {
	file := flag.String("config", "dev.env", "config file")
	flag.Parse()

	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	logger.Log("msg", "migrate starting...")
	defer logger.Log("msg", "migrate shut down")

	cfg, err := app.LoadConfig(*file)
	if err != nil {
		logger.Log("msg", "loading config", "err", err)
		return
	}
	logger.Log("msg", "config loaded")

	pgURL := composePostgresURL(cfg.Postgres)
	mg, err := migrate.New("file://db/migrations", pgURL)
	if err != nil {
		logger.Log("msg", "initializing migrations", "err", err)
		return
	}
	if err := mg.Up(); err != nil {
		logger.Log("msg", "running migrations up", "err", err)
		return
	}
	logger.Log("msg", "migrate done ok")
}

func composePostgresURL(cfg app.PostgresConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?connect_timeout=%d&application_name=%s&sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Timeout, "migrate-dummy-go",
	)
}
