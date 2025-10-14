package service_test

import (
	"testing"

	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
	"github.com/NiClassic/go-cloud/internal/testutil"
)

func setupFolderTest(t *testing.T) (*service.FolderService, int64, string, string) {
	t.Helper()

	db := testutil.SetupTestDB(t)
	ctx := testutil.TestContext(t)
	tmpDir := testutil.SetupTestStorage(t)

	userRepo := repository.NewUserRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	fileRepo := repository.NewPersonalFileRepository(db)
	st := storage.NewIOStorage(tmpDir)

	folderSvc := service.NewFolderService(folderRepo, fileRepo, st)

	// Create a test user
	userID, err := userRepo.Insert(ctx, "testuser", "hashedpass")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return folderSvc, userID, "testuser", tmpDir
}

func TestFolderService_CreateFolder(t *testing.T) {
	folderSvc, userID, username, _ := setupFolderTest(t)
	ctx := testutil.TestContext(t)

	tests := []struct {
		name        string
		folderName  string
		parentID    int64
		path        string
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "create root folder",
			folderName: username,
			parentID:   -1,
			path:       "/",
			wantErr:    false,
		},
		{
			name:       "create subfolder",
			folderName: "documents",
			parentID:   -1,
			path:       "/documents",
			wantErr:    false,
		},
		{
			name:        "empty folder name",
			folderName:  "",
			parentID:    -1,
			path:        "/",
			wantErr:     true,
			expectedErr: service.ErrInvalidFolderName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folder, err := folderSvc.CreateFolder(ctx, userID, username, tt.parentID, tt.folderName, tt.path)

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

			if folder == nil {
				t.Fatal("expected folder but got nil")
			}

			if folder.Name != tt.folderName {
				t.Errorf("expected name %s, got %s", tt.folderName, folder.Name)
			}

			if folder.UserID != userID {
				t.Errorf("expected user ID %d, got %d", userID, folder.UserID)
			}
		})
	}
}

func TestFolderService_GetById(t *testing.T) {
	folderSvc, userID, username, _ := setupFolderTest(t)
	ctx := testutil.TestContext(t)

	// Create a test folder
	folder, err := folderSvc.CreateFolder(ctx, userID, username, -1, "testfolder", "/testfolder")
	if err != nil {
		t.Fatalf("failed to create test folder: %v", err)
	}

	t.Run("valid folder ID", func(t *testing.T) {
		retrieved, err := folderSvc.GetById(ctx, userID, folder.ID)
		if err != nil {
			t.Fatalf("failed to get folder: %v", err)
		}

		if retrieved.ID != folder.ID {
			t.Errorf("expected ID %d, got %d", folder.ID, retrieved.ID)
		}

		if retrieved.Name != folder.Name {
			t.Errorf("expected name %s, got %s", folder.Name, retrieved.Name)
		}
	})

	t.Run("wrong user ID", func(t *testing.T) {
		_, err := folderSvc.GetById(ctx, userID+999, folder.ID)
		if err != service.ErrFolderNotFound {
			t.Errorf("expected ErrFolderNotFound, got %v", err)
		}
	})

	t.Run("non-existent folder", func(t *testing.T) {
		_, err := folderSvc.GetById(ctx, userID, 99999)
		if err == nil {
			t.Error("expected error for non-existent folder")
		}
	})
}

func TestFolderService_GetByPath(t *testing.T) {
	folderSvc, userID, username, _ := setupFolderTest(t)
	ctx := testutil.TestContext(t)

	// Create test folders
	root, err := folderSvc.CreateFolder(ctx, userID, username, -1, username, "/")
	if err != nil {
		t.Fatalf("failed to create root folder: %v", err)
	}

	_, err = folderSvc.CreateFolder(ctx, userID, username, root.ID, "documents", "/documents")
	if err != nil {
		t.Fatalf("failed to create documents folder: %v", err)
	}

	tests := []struct {
		name       string
		path       string
		expectName string
		wantErr    bool
	}{
		{
			name:       "root folder",
			path:       "/",
			expectName: username,
			wantErr:    false,
		},
		{
			name:       "subfolder",
			path:       "/documents",
			expectName: "documents",
			wantErr:    false,
		},
		{
			name:    "non-existent path",
			path:    "/nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folder, err := folderSvc.GetByPath(ctx, userID, username, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if folder.Name != tt.expectName {
				t.Errorf("expected name %s, got %s", tt.expectName, folder.Name)
			}
		})
	}
}

func TestFolderService_GetFolderContents(t *testing.T) {
	folderSvc, userID, username, _ := setupFolderTest(t)
	ctx := testutil.TestContext(t)

	// Create folder structure
	root, err := folderSvc.CreateFolder(ctx, userID, username, -1, username, "/")
	if err != nil {
		t.Fatalf("failed to create root folder: %v", err)
	}

	// Create subfolders
	_, err = folderSvc.CreateFolder(ctx, userID, username, root.ID, "documents", "/documents")
	if err != nil {
		t.Fatalf("failed to create documents folder: %v", err)
	}

	_, err = folderSvc.CreateFolder(ctx, userID, username, root.ID, "photos", "/photos")
	if err != nil {
		t.Fatalf("failed to create photos folder: %v", err)
	}

	folders, files, err := folderSvc.GetFolderContents(ctx, userID, root.ID)
	if err != nil {
		t.Fatalf("failed to get folder contents: %v", err)
	}

	if len(folders) != 2 {
		t.Errorf("expected 2 folders, got %d", len(folders))
	}

	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}

	// Verify folder names
	folderNames := make(map[string]bool)
	for _, f := range folders {
		folderNames[f.Name] = true
	}

	if !folderNames["documents"] || !folderNames["photos"] {
		t.Error("expected to find documents and photos folders")
	}
}

func TestFolderService_MoveFolder(t *testing.T) {
	folderSvc, userID, username, _ := setupFolderTest(t)
	ctx := testutil.TestContext(t)

	// Create folder structure
	root, _ := folderSvc.CreateFolder(ctx, userID, username, -1, username, "/")
	docs, _ := folderSvc.CreateFolder(ctx, userID, username, root.ID, "documents", "/documents")
	work, _ := folderSvc.CreateFolder(ctx, userID, username, root.ID, "work", "/work")

	t.Run("move folder to new parent", func(t *testing.T) {
		err := folderSvc.MoveFolder(ctx, userID, docs.ID, work.ID)
		if err != nil {
			t.Fatalf("failed to move folder: %v", err)
		}

		// Verify the move
		folders, _, err := folderSvc.GetFolderContents(ctx, userID, work.ID)
		if err != nil {
			t.Fatalf("failed to get folder contents: %v", err)
		}

		if len(folders) != 1 {
			t.Errorf("expected 1 folder in work, got %d", len(folders))
		}

		if len(folders) > 0 && folders[0].Name != "documents" {
			t.Errorf("expected documents folder, got %s", folders[0].Name)
		}
	})

	t.Run("move non-existent folder", func(t *testing.T) {
		err := folderSvc.MoveFolder(ctx, userID, 99999, work.ID)
		if err == nil {
			t.Error("expected error for non-existent folder")
		}
	})
}

func TestFolderService_DeleteFolder(t *testing.T) {
	folderSvc, userID, username, _ := setupFolderTest(t)
	ctx := testutil.TestContext(t)

	// Create a test folder
	root, _ := folderSvc.CreateFolder(ctx, userID, username, -1, username, "/")
	folder, err := folderSvc.CreateFolder(ctx, userID, username, root.ID, "todelete", "/todelete")
	if err != nil {
		t.Fatalf("failed to create test folder: %v", err)
	}

	t.Run("delete folder", func(t *testing.T) {
		err := folderSvc.DeleteFolder(ctx, userID, folder.ID)
		if err != nil {
			t.Fatalf("failed to delete folder: %v", err)
		}

		_, err = folderSvc.GetById(ctx, userID, folder.ID)
		if err == nil {
			t.Error("expected error when getting deleted folder")
		}
	})

	t.Run("delete non-existent folder", func(t *testing.T) {
		err := folderSvc.DeleteFolder(ctx, userID, 99999)
		if err == nil {
			t.Error("expected error for non-existent folder")
		}
	})

	t.Run("delete with wrong user", func(t *testing.T) {
		folder2, _ := folderSvc.CreateFolder(ctx, userID, username, root.ID, "another", "/another")
		err := folderSvc.DeleteFolder(ctx, userID+999, folder2.ID)
		if err != service.ErrFolderNotFound {
			t.Errorf("expected ErrFolderNotFound, got %v", err)
		}
	})
}
