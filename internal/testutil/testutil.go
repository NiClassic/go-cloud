package testutil

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/NiClassic/go-cloud/internal/db"
	_ "modernc.org/sqlite"
)

func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := testDB.Ping(); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	migrationsDir := filepath.Join("../..", "db", "migrations")
	if err := db.Migrate(testDB, "file://"+migrationsDir); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	t.Cleanup(func() {
		if err := testDB.Close(); err != nil {
			t.Errorf("failed to close test database: %v", err)
		}
	})

	return testDB
}

func SetupTestStorage(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "go-cloud-test-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("failed to remove temp directory: %v", err)
		}
	})

	return tmpDir
}

func TestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return ctx
}

func CreateTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	filePath := filepath.Join(dir, name)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	return filePath
}
