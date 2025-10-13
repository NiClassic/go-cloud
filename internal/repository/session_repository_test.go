package repository_test

import (
	"testing"

	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/testutil"
)

func setupSessionRepositoryTest(t *testing.T) (*repository.SessionRepository, *repository.UserRepository) {
	t.Helper()

	db := testutil.SetupTestDB(t)

	sessionRepo := repository.NewSessionRepository(db)
	userRepo := repository.NewUserRepository(db)
	return sessionRepo, userRepo
}

func TestSessionRepository_Insert(t *testing.T) {
	sessionRepo, userRepo := setupSessionRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, err := userRepo.Insert(ctx, "bob", "hashedpass")
	if err != nil {
		t.Fatal("could not insert test user")
	}

	tests := []struct {
		name         string
		userID       int64
		sessionToken string
		wantErr      bool
	}{
		{
			name:         "valid session",
			userID:       userID,
			sessionToken: "token123",
			wantErr:      false,
		},
		{
			name:         "another valid session",
			userID:       userID,
			sessionToken: "token456",
			wantErr:      false,
		},
		{
			name:         "duplicate token",
			userID:       userID,
			sessionToken: "token123",
			wantErr:      true,
		},
		{
			name:         "invalid user ID",
			userID:       99999,
			sessionToken: "token789",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionID, err := sessionRepo.Insert(ctx, tt.userID, tt.sessionToken)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if sessionID <= 0 {
				t.Error("expected valid session ID")
			}
		})
	}
}

func TestSessionRepository_GetByToken(t *testing.T) {
	sessionRepo, userRepo := setupSessionRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, err := userRepo.Insert(ctx, "bob", "hashedpass")
	if err != nil {
		t.Fatal("could not insert test user")
	}

	sessionToken := "testtoken123"
	sessionID, err := sessionRepo.Insert(ctx, userID, sessionToken)
	if err != nil {
		t.Fatal("could not insert test session")
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "existing token",
			token:   sessionToken,
			wantErr: false,
		},
		{
			name:    "non-existent token",
			token:   "nonexistent",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := sessionRepo.GetByToken(ctx, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if session.ID != sessionID {
				t.Errorf("expected session ID %d, got %d", sessionID, session.ID)
			}

			if session.UserID != userID {
				t.Errorf("expected user ID %d, got %d", userID, session.UserID)
			}

			if session.SessionToken != tt.token {
				t.Errorf("expected token %s, got %s", tt.token, session.SessionToken)
			}

			if !session.Valid {
				t.Error("expected session to be valid")
			}

			if session.CreatedAt.IsZero() {
				t.Error("expected non-zero created_at timestamp")
			}
		})
	}
}

func TestSessionRepository_Invalidate(t *testing.T) {
	sessionRepo, userRepo := setupSessionRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, err := userRepo.Insert(ctx, "bob", "hashedpass")
	if err != nil {
		t.Fatal("could not insert test user")
	}

	sessionToken := "testtoken123"
	_, err = sessionRepo.Insert(ctx, userID, sessionToken)
	if err != nil {
		t.Fatal("could not insert test session")
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "invalidate existing session",
			token:   sessionToken,
			wantErr: false,
		},
		// Update with zero target does not return an error
		{
			name:    "invalidate non-existent session",
			token:   "nonexistent",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sessionRepo.Invalidate(ctx, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.token == sessionToken {
				session, err := sessionRepo.GetByToken(ctx, tt.token)
				if err != nil {
					t.Fatalf("could not get session: %v", err)
				}

				if session.Valid {
					t.Error("expected session to be invalid")
				}
			}
		})
	}
}

func TestSessionRepository_DeleteByUser(t *testing.T) {
	sessionRepo, userRepo := setupSessionRepositoryTest(t)
	ctx := testutil.TestContext(t)

	user1ID, _ := userRepo.Insert(ctx, "user1", "pass1")
	user2ID, _ := userRepo.Insert(ctx, "user2", "pass2")

	tokens1 := []string{"token1", "token2", "token3"}
	for _, token := range tokens1 {
		_, err := sessionRepo.Insert(ctx, user1ID, token)
		if err != nil {
			t.Fatal("could not insert test session")
		}
	}

	token2 := "user2token"
	sessionRepo.Insert(ctx, user2ID, token2)

	tests := []struct {
		name    string
		userID  int64
		wantErr bool
	}{
		{
			name:    "delete all sessions for user1",
			userID:  user1ID,
			wantErr: false,
		},
		// Delete with 0 targets is not an error
		{
			name:    "delete sessions for non-existent user",
			userID:  99999,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sessionRepo.DeleteByUser(ctx, tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.userID == user1ID {
				for _, token := range tokens1 {
					_, err := sessionRepo.GetByToken(ctx, token)
					if err == nil {
						t.Errorf("expected session %s to be deleted", token)
					}
				}

				session, err := sessionRepo.GetByToken(ctx, token2)
				if err != nil {
					t.Fatal("user2's session should still exist")
				}
				if session.UserID != user2ID {
					t.Error("user2's session was affected")
				}
			}
		})
	}
}

func TestSessionRepository_MultipleSessionsPerUser(t *testing.T) {
	sessionRepo, userRepo := setupSessionRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, err := userRepo.Insert(ctx, "bob", "hashedpass")
	if err != nil {
		t.Fatal("could not insert test user")
	}

	tokens := []string{"session1", "session2", "session3"}
	for _, token := range tokens {
		_, err := sessionRepo.Insert(ctx, userID, token)
		if err != nil {
			t.Fatalf("could not insert session %s: %v", token, err)
		}
	}

	for _, token := range tokens {
		session, err := sessionRepo.GetByToken(ctx, token)
		if err != nil {
			t.Fatalf("could not get session %s: %v", token, err)
		}

		if session.UserID != userID {
			t.Errorf("session %s has wrong user ID", token)
		}

		if session.SessionToken != token {
			t.Errorf("expected token %s, got %s", token, session.SessionToken)
		}
	}
}

func TestSessionRepository_TokenUniqueness(t *testing.T) {
	sessionRepo, userRepo := setupSessionRepositoryTest(t)
	ctx := testutil.TestContext(t)

	user1ID, _ := userRepo.Insert(ctx, "user1", "pass1")
	user2ID, _ := userRepo.Insert(ctx, "user2", "pass2")

	token := "uniquetoken"

	_, err := sessionRepo.Insert(ctx, user1ID, token)
	if err != nil {
		t.Fatalf("could not insert first session: %v", err)
	}

	_, err = sessionRepo.Insert(ctx, user2ID, token)
	if err == nil {
		t.Fatal("expected error when inserting duplicate token, got none")
	}
}

func TestSessionRepository_CascadeDelete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := testutil.TestContext(t)

	sessionRepo := repository.NewSessionRepository(db)
	userRepo := repository.NewUserRepository(db)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
	token := "testtoken"
	sessionRepo.Insert(ctx, userID, token)

	_, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		t.Fatalf("could not delete user: %v", err)
	}

	_, err = sessionRepo.GetByToken(ctx, token)
	if err == nil {
		t.Error("expected session to be deleted via cascade")
	}
}

func TestSessionRepository_InvalidateIdempotent(t *testing.T) {
	sessionRepo, userRepo := setupSessionRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
	token := "testtoken"
	sessionRepo.Insert(ctx, userID, token)

	err := sessionRepo.Invalidate(ctx, token)
	if err != nil {
		t.Fatalf("first invalidation failed: %v", err)
	}

	err = sessionRepo.Invalidate(ctx, token)
	if err != nil {
		t.Fatalf("second invalidation failed: %v", err)
	}

	session, err := sessionRepo.GetByToken(ctx, token)
	if err != nil {
		t.Fatalf("could not get session: %v", err)
	}

	if session.Valid {
		t.Error("session should remain invalid after multiple invalidations")
	}
}

func TestSessionRepository_GetByToken_ReturnsCorrectFields(t *testing.T) {
	sessionRepo, userRepo := setupSessionRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
	token := "testtoken"
	sessionID, _ := sessionRepo.Insert(ctx, userID, token)

	session, err := sessionRepo.GetByToken(ctx, token)
	if err != nil {
		t.Fatalf("could not get session: %v", err)
	}

	if session.ID != sessionID {
		t.Errorf("expected ID %d, got %d", sessionID, session.ID)
	}

	if session.UserID != userID {
		t.Errorf("expected user ID %d, got %d", userID, session.UserID)
	}

	if session.SessionToken != token {
		t.Errorf("expected token %s, got %s", token, session.SessionToken)
	}

	if !session.Valid {
		t.Error("expected new session to be valid")
	}

	if session.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}
