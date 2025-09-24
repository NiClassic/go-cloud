package main

import (
	"database/sql"
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
	"net/http"
	"os"
)

func main() {
	config.Init()

	logger.Init(config.Debug)

	err := godotenv.Load(".env")
	if err != nil {
		logger.Info("did not find .env file, falling back to shell environment")
	}

	timezoneName := os.Getenv("TZ")
	if timezoneName == "" {
		timezoneName = "UTC"
		logger.Info("did not find TZ environment variable, using UTC")
	}

	if err := timezone.Init(timezoneName); err != nil {
		logger.Fatal("could not initialize timezone '%s': %v", timezoneName, err)
	}

	dbConn, err := db.New()
	if err != nil {
		logger.Fatal("could not connect to db: %v", err)
	}

	if err := db.Migrate(dbConn); err != nil {
		logger.Fatal("could not apply migrations: %v", err)
	}

	defer func(db *sql.DB) {
		if err = db.Close(); err != nil {
			logger.Fatal("could not close database: %v", err)
		}
	}(dbConn)

	st := storage.NewIOStorage(os.Getenv("DATA_ROOT"))

	services := service.InitServices(dbConn, st)

	tmpl, err := handler.ParseTemplates()
	if err != nil {
		logger.Fatal("could not parse templates: %v", err)
	}

	mux := handler.New(services, st, tmpl)

	logger.Info("timezone is %s", timezoneName)
	logger.Info("listening on :8080 (Debug Mode=%v)", config.Debug)
	if err = http.ListenAndServe(":8080", mux); err != nil {
		logger.Fatal("could not run server: %v", err)
	}
}
