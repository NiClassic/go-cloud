package repository_test

import (
	"testing"

	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/testutil"
)

func setupUserRepositoryTest(t *testing.T) *repository.UserRepository {
	t.Helper()

	db := testutil.SetupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	return userRepo
}

func TestUserRepository_Insert(t *testing.T) {
	repo := setupUserRepositoryTest(t)
	ctx := testutil.TestContext(t)

	tests := []struct {
		name           string
		username       string
		hashedPassword string
		wantID         int64
		wantErr        bool
	}{
		{"valid user", "bob", "hashed_password", 1, false},
		{"unique constraint on username", "bob", "hashed_password", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := repo.Insert(ctx, tt.username, tt.hashedPassword)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if id <= 0 {
				t.Fatalf("expected valid user ID, got %v", id)
			}

			if tt.wantID != id {
				t.Fatalf("expected user ID %v, got %v", tt.wantID, id)
			}
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	repo := setupUserRepositoryTest(t)
	ctx := testutil.TestContext(t)

	id, err := repo.Insert(ctx, "bob", "hashed_password")
	if err != nil {
		t.Fatal("could not insert test user")
	}

	tests := []struct {
		name     string
		username string
		wantID   int64
		wantErr  bool
	}{
		{"valid user", "bob", 1, false},
		{"unknown user name", "not_existing", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByUsername(ctx, tt.username)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if tt.wantID != id {
				t.Fatalf("expected user ID %v, got %v", tt.wantID, id)
			}

			if tt.username != user.Username {
				t.Fatalf("expected username %v, got %v", tt.username, user.Username)
			}

		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	repo := setupUserRepositoryTest(t)
	ctx := testutil.TestContext(t)

	id, err := repo.Insert(ctx, "bob", "hashed_password")
	if err != nil {
		t.Fatal("could not insert test user")
	}

	tests := []struct {
		name         string
		ID           int64
		wantUsername string
		wantErr      bool
	}{
		{"valid user", 1, "bob", false},
		{"unknown user name", 9999, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByID(ctx, tt.ID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if tt.wantUsername != user.Username {
				t.Fatalf("expected username %v, got %v", tt.wantUsername, user.Username)
			}

			if tt.ID != user.ID {
				t.Fatalf("expected user ID %v, got %v", tt.wantUsername, id)
			}
		})
	}
}
