package service_test

import (
	"testing"

	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/testutil"
)

func TestAuthService_Register(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := testutil.TestContext(t)

	userRepo := repository.NewUserRepository(db)
	sessRepo := repository.NewSessionRepository(db)
	authSvc := service.NewAuthService(userRepo, sessRepo)

	tests := []struct {
		name        string
		username    string
		password    string
		wantErr     bool
		expectedErr error
	}{
		{
			"valid registration",
			"bob",
			"password",
			false,
			nil,
		},
		{
			"empty username",
			"",
			"password",
			true,
			service.ErrEmptyCredentials,
		},
		{
			"empty password",
			"bob",
			"",
			true,
			service.ErrEmptyCredentials,
		},
		{
			"both empty",
			"",
			"",
			true,
			service.ErrEmptyCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := authSvc.Register(ctx, tt.username, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Fatal("unexpected error but gone")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if userID <= 0 {
				t.Error("expected valid user ID")
			}
		})
	}
}

func TestAuthService_Authenticate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := testutil.TestContext(t)

	userRepo := repository.NewUserRepository(db)
	sessRepo := repository.NewSessionRepository(db)
	authSvc := service.NewAuthService(userRepo, sessRepo)

	username := "testuser"
	password := "testpass123"
	_, err := authSvc.Register(ctx, username, password)
	if err != nil {
		t.Fatalf("failed to setup test user: %v", err)
	}

	tests := []struct {
		name        string
		username    string
		password    string
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "valid credentials",
			username: username,
			password: password,
			wantErr:  false,
		},
		{
			name:        "wrong password",
			username:    username,
			password:    "wrongpass",
			wantErr:     true,
			expectedErr: service.ErrInvalidCredentials,
		},
		{
			name:        "non-existent user",
			username:    "nonexistent",
			password:    password,
			wantErr:     true,
			expectedErr: service.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := authSvc.Authenticate(ctx, tt.username, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user == nil {
				t.Fatal("expected user but got nil")
			}

			if user.Username != tt.username {
				t.Errorf("expected username %s, got %s", tt.username, user.Username)
			}
		})
	}
}

func TestAuthService_SessionManagement(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := testutil.TestContext(t)

	userRepo := repository.NewUserRepository(db)
	sessRepo := repository.NewSessionRepository(db)
	authSvc := service.NewAuthService(userRepo, sessRepo)

	username := "testuser"
	password := "testpass123"
	_, err := authSvc.Register(ctx, username, password)
	if err != nil {
		t.Fatalf("failed to setup test user: %v", err)
	}

	user, err := authSvc.Authenticate(ctx, username, password)
	if err != nil {
		t.Fatalf("failed to authenticate test user: %v", err)
	}

	t.Run("register session", func(t *testing.T) {
		token, err := authSvc.RegisterSession(ctx, user)
		if err != nil {
			t.Fatalf("failed to register session: %v", err)
		}

		if token == "" {
			t.Error("expected non-empty token")
		}

		if len(token) != 64 { // 32 bytes hex encoded
			t.Errorf("expected token length 64, got %d", len(token))
		}
	})

	t.Run("validate session", func(t *testing.T) {
		token, _ := authSvc.RegisterSession(ctx, user)

		valid, err := authSvc.ValidateSession(ctx, token)
		if err != nil {
			t.Fatalf("failed to validate session: %v", err)
		}

		if !valid {
			t.Error("expected session to be valid")
		}
	})

	t.Run("get user by session token", func(t *testing.T) {
		token, _ := authSvc.RegisterSession(ctx, user)

		retrievedUser, err := authSvc.GetUserBySessionToken(ctx, token)
		if err != nil {
			t.Fatalf("failed to get user by token: %v", err)
		}

		if retrievedUser.ID != user.ID {
			t.Errorf("expected user ID %d, got %d", user.ID, retrievedUser.ID)
		}

		if retrievedUser.Username != user.Username {
			t.Errorf("expected username %s, got %s", user.Username, retrievedUser.Username)
		}
	})

	t.Run("destroy session", func(t *testing.T) {
		token, _ := authSvc.RegisterSession(ctx, user)

		err := authSvc.DestroySession(ctx, token)
		if err != nil {
			t.Fatalf("failed to destroy session: %v", err)
		}

		valid, err := authSvc.ValidateSession(ctx, token)
		if err != nil {
			t.Fatalf("failed to validate destroyed session: %v", err)
		}

		if valid {
			t.Error("expected session to be invalid after destruction")
		}
	})

	t.Run("invalid session token", func(t *testing.T) {
		_, err := authSvc.GetUserBySessionToken(ctx, "invalid-token")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})
}
