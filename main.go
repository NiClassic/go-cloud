package main

import (
	"database/sql"
	"errors"
	"github.com/NiClassic/go-cloud/config"
	"github.com/NiClassic/go-cloud/internal/handler"
	"github.com/NiClassic/go-cloud/internal/logger"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
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
		logger.Fatal("could not load .env file: %v", err)
	}
	db, err := sql.Open("sqlite", os.Getenv("DB_FILE"))
	if err != nil {
		logger.Fatal("could not open database: %v", err)
	}
	if err = db.Ping(); err != nil {
		logger.Fatal("could not ping database: %v", err)
	}
	logger.Info("successfully connected to database")

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		logger.Fatal("could not open database for migrations: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://./db/migrations", "sqlite", driver)
	if err != nil {
		logger.Fatal("could not initialize migrations: %v", err)
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Fatal("could not run migrations: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatal("could not close database: %v", err)
		}
	}(db)

	storage := storage.NewStorage(os.Getenv("DATA_ROOT"))

	services := service.InitServices(db, storage)

	tmpl, err := handler.ParseTemplates()
	if err != nil {
		logger.Fatal("could not parse templates: %v", err)
	}

	mux := handler.New(services, storage, tmpl)

	logger.Info("listening on :8080 (Debug Mode=%v)", config.Debug)
	if err = http.ListenAndServe(":8080", mux); err != nil {
		logger.Fatal("could not run server: %v", err)
	}
}
