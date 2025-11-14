package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/NiClassic/go-cloud/config"
	"github.com/NiClassic/go-cloud/internal/db"
	"github.com/NiClassic/go-cloud/internal/handler"
	"github.com/NiClassic/go-cloud/internal/logger"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
	"github.com/NiClassic/go-cloud/internal/timezone"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		logger.Info("did not find .env file, falling back to shell environment")
	}
	cfg := config.Init()
	logger.Init(cfg.DebugMode)

	if err := timezone.Init(cfg.TimezoneName); err != nil {
		logger.Fatal("could not initialize timezone '%s': %v", cfg.TimezoneName, err)
	}

	dbConn, err := db.New()
	if err != nil {
		logger.Fatal("could not connect to db: %v", err)
	}

	if err := db.Migrate(dbConn, "file://./db/migrations"); err != nil {
		logger.Fatal("could not apply migrations: %v", err)
	}

	defer func(db *sql.DB) {
		if err = db.Close(); err != nil {
			logger.Fatal("could not close database: %v", err)
		}
	}(dbConn)

	st := storage.NewIOStorage(os.Getenv("DATA_ROOT"))

	services := service.InitServices(dbConn, st)
	renderer, err := handler.NewRenderer(cfg)
	if err != nil {
		logger.Fatal("could not initialize renderer: %v", err)
	}

	mux := handler.New(cfg, renderer, services, st)

	logger.Info("timezone is %s", cfg.TimezoneName)
	logger.Info("listening on :8080 (Debug Mode=%v)", cfg.DebugMode)
	if err = http.ListenAndServe(":8080", mux); err != nil {
		logger.Fatal("could not run server: %v", err)
	}
}
