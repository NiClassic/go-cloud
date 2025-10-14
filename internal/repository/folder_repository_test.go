package repository_test

import (
	"testing"

	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/testutil"
)

func setupFolderRepositoryTest(t *testing.T) (*repository.FolderRepository, *repository.UserRepository) {
	t.Helper()

	db := testutil.SetupTestDB(t)

	folderRepo := repository.NewFolderRepository(db)
	userRepo := repository.NewUserRepository(db)
	return folderRepo, userRepo
}

func TestFolderRepository_Insert(t *testing.T) {
	folderRepo, userRepo := setupFolderRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, err := userRepo.Insert(ctx, "bob", "hashedpass")
	if err != nil {
		t.Fatal("could not insert test user")
	}

	var testID int64 = 1

	tests := []struct {
		name     string
		userID   int64
		parentID *int64
		fname    string
		path     string
		wantErr  bool
	}{
		{
			name:     "valid root folder",
			userID:   userID,
			parentID: nil,
			fname:    "bob",
			path:     "bob",
			wantErr:  false,
		},
		{
			name:     "valid subfolder",
			userID:   userID,
			parentID: &testID,
			fname:    "documents",
			path:     "bob/documents",
			wantErr:  false,
		},
		{
			name:     "duplicate path",
			userID:   userID,
			parentID: nil,
			fname:    "bob",
			path:     "bob",
			wantErr:  true,
		},
		{
			name:     "invalid user ID",
			userID:   99999,
			parentID: nil,
			fname:    "folder",
			path:     "invalid/folder",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folderID, err := folderRepo.Insert(ctx, tt.userID, tt.parentID, tt.fname, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if folderID <= 0 {
				t.Error("expected valid folder ID")
			}
		})
	}
}

func TestFolderRepository_GetByID(t *testing.T) {
	folderRepo, userRepo := setupFolderRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, err := userRepo.Insert(ctx, "bob", "hashedpass")
	if err != nil {
		t.Fatal("could not insert test user")
	}

	folderName := "testfolder"
	folderPath := "bob/testfolder"
	folderID, err := folderRepo.Insert(ctx, userID, nil, folderName, folderPath)
	if err != nil {
		t.Fatal("could not insert test folder")
	}

	tests := []struct {
		name     string
		folderID int64
		wantErr  bool
	}{
		{
			name:     "existing folder",
			folderID: folderID,
			wantErr:  false,
		},
		{
			name:     "non-existent folder",
			folderID: 99999,
			wantErr:  true,
		},
		{
			name:     "zero folder ID",
			folderID: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folder, err := folderRepo.GetByID(ctx, tt.folderID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if folder.ID != folderID {
				t.Errorf("expected folder ID %d, got %d", folderID, folder.ID)
			}

			if folder.Name != folderName {
				t.Errorf("expected folder name %s, got %s", folderName, folder.Name)
			}

			if folder.Path != folderPath {
				t.Errorf("expected path %s, got %s", folderPath, folder.Path)
			}

			if folder.UserID != userID {
				t.Errorf("expected user ID %d, got %d", userID, folder.UserID)
			}

			if folder.CreatedAt.IsZero() {
				t.Error("expected non-zero created_at timestamp")
			}

			if folder.UpdatedAt.IsZero() {
				t.Error("expected non-zero updated_at timestamp")
			}
		})
	}
}

func TestFolderRepository_GetByPath(t *testing.T) {
	folderRepo, userRepo := setupFolderRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, err := userRepo.Insert(ctx, "bob", "hashedpass")
	if err != nil {
		t.Fatal("could not insert test user")
	}

	path1 := "bob"
	path2 := "bob/documents"

	folderID1, _ := folderRepo.Insert(ctx, userID, nil, "bob", path1)
	folderRepo.Insert(ctx, userID, &folderID1, "documents", path2)

	tests := []struct {
		name     string
		path     string
		wantName string
		wantErr  bool
	}{
		{
			name:     "root folder",
			path:     path1,
			wantName: "bob",
			wantErr:  false,
		},
		{
			name:     "subfolder",
			path:     path2,
			wantName: "documents",
			wantErr:  false,
		},
		{
			name:    "non-existent path",
			path:    "nonexistent",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folder, err := folderRepo.GetByPath(ctx, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if folder.Name != tt.wantName {
				t.Errorf("expected folder name %s, got %s", tt.wantName, folder.Name)
			}

			if folder.Path != tt.path {
				t.Errorf("expected path %s, got %s", tt.path, folder.Path)
			}
		})
	}
}

func TestFolderRepository_GetByUser(t *testing.T) {
	folderRepo, userRepo := setupFolderRepositoryTest(t)
	ctx := testutil.TestContext(t)

	user1ID, _ := userRepo.Insert(ctx, "user1", "pass1")
	user2ID, _ := userRepo.Insert(ctx, "user2", "pass2")

	rootID1, _ := folderRepo.Insert(ctx, user1ID, nil, "user1", "user1")
	rootID2, _ := folderRepo.Insert(ctx, user2ID, nil, "user2", "user2")

	folderRepo.Insert(ctx, user1ID, &rootID1, "docs", "user1/docs")
	folderRepo.Insert(ctx, user2ID, &rootID2, "pics", "user2/pics")

	folderRepo.Insert(ctx, user1ID, nil, "folder2", "user1/folder2")
	folderRepo.Insert(ctx, user1ID, nil, "folder3", "user1/folder3")

	tests := []struct {
		name          string
		userID        int64
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "user1 folders",
			userID:        user1ID,
			expectedCount: 3,
			wantErr:       false,
		},
		{
			name:          "user2 folders",
			userID:        user2ID,
			expectedCount: 1,
			wantErr:       false,
		},
		{
			name:          "non-existent user",
			userID:        99999,
			expectedCount: 0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folders, err := folderRepo.GetByUser(ctx, tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(folders) != tt.expectedCount {
				t.Errorf("expected %d folders, got %d", tt.expectedCount, len(folders))
			}

			for _, folder := range folders {
				if folder.UserID != tt.userID {
					t.Errorf("folder %d belongs to user %d, expected %d", folder.ID, folder.UserID, tt.userID)
				}
			}
		})
	}
}

func TestFolderRepository_GetByUserAndParent(t *testing.T) {
	folderRepo, userRepo := setupFolderRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")

	rootID, _ := folderRepo.Insert(ctx, userID, nil, "bob", "bob")
	folderRepo.Insert(ctx, userID, &rootID, "docs", "bob/docs")
	folderRepo.Insert(ctx, userID, &rootID, "pics", "bob/pics")
	folderRepo.Insert(ctx, userID, &rootID, "videos", "bob/videos")

	tests := []struct {
		name          string
		userID        int64
		parentID      int64
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "subfolders of root",
			userID:        userID,
			parentID:      rootID,
			expectedCount: 3,
			wantErr:       false,
		},
		{
			name:          "wrong user ID",
			userID:        99999,
			parentID:      rootID,
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name:          "non-existent parent",
			userID:        userID,
			parentID:      99999,
			expectedCount: 0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folders, err := folderRepo.GetByUserAndParent(ctx, tt.userID, tt.parentID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(folders) != tt.expectedCount {
				t.Errorf("expected %d folders, got %d", tt.expectedCount, len(folders))
			}

			for _, folder := range folders {
				if folder.UserID != tt.userID {
					t.Errorf("folder belongs to user %d, expected %d", folder.UserID, tt.userID)
				}

				if !folder.ParentID.Valid || folder.ParentID.Int64 != tt.parentID {
					t.Errorf("folder has parent %v, expected %d", folder.ParentID, tt.parentID)
				}
			}
		})
	}
}

func TestFolderRepository_Delete(t *testing.T) {
	folderRepo, userRepo := setupFolderRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
	folderID, err := folderRepo.Insert(ctx, userID, nil, "todelete", "bob/todelete")
	if err != nil {
		t.Fatal("could not insert test folder")
	}

	tests := []struct {
		name     string
		folderID int64
		wantErr  bool
	}{
		{
			name:     "delete existing folder",
			folderID: folderID,
			wantErr:  false,
		},
		// Delete zero rows is not an error
		{
			name:     "delete non-existent folder",
			folderID: 99999,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := folderRepo.Delete(ctx, tt.folderID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.folderID == folderID {
				_, err := folderRepo.GetByID(ctx, tt.folderID)
				if err == nil {
					t.Error("expected error when getting deleted folder")
				}
			}
		})
	}
}

func TestFolderRepository_UpdatePath(t *testing.T) {
	folderRepo, userRepo := setupFolderRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
	folderID, err := folderRepo.Insert(ctx, userID, nil, "folder", "bob/folder")
	if err != nil {
		t.Fatal("could not insert test folder")
	}

	tests := []struct {
		name     string
		folderID int64
		newPath  string
		wantErr  bool
	}{
		{
			name:     "valid path update",
			folderID: folderID,
			newPath:  "bob/renamed",
			wantErr:  false,
		},
		// Update with no targets is not an error
		{
			name:     "update non-existent folder",
			folderID: 99999,
			newPath:  "bob/nonexistent",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := folderRepo.UpdatePath(ctx, tt.folderID, tt.newPath)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.folderID == folderID {
				folder, err := folderRepo.GetByID(ctx, tt.folderID)
				if err != nil {
					t.Fatalf("could not get updated folder: %v", err)
				}

				if folder.Path != tt.newPath {
					t.Errorf("expected path %s, got %s", tt.newPath, folder.Path)
				}
			}
		})
	}
}

func TestFolderRepository_UpdateParent(t *testing.T) {
	folderRepo, userRepo := setupFolderRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
	parentID1, _ := folderRepo.Insert(ctx, userID, nil, "parent1", "bob/parent1")
	parentID2, _ := folderRepo.Insert(ctx, userID, nil, "parent2", "bob/parent2")
	childID, _ := folderRepo.Insert(ctx, userID, &parentID1, "child", "bob/parent1/child")

	tests := []struct {
		name        string
		folderID    int64
		newParentID *int64
		wantErr     bool
	}{
		{
			name:        "move to different parent",
			folderID:    childID,
			newParentID: &parentID2,
			wantErr:     false,
		},
		{
			name:        "update non-existent folder",
			folderID:    99999,
			newParentID: &parentID1,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := folderRepo.UpdateParent(ctx, tt.folderID, *(tt.newParentID))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.folderID == childID {
				folder, err := folderRepo.GetByID(ctx, tt.folderID)
				if err != nil {
					t.Fatalf("could not get updated folder: %v", err)
				}

				if !folder.ParentID.Valid || folder.ParentID.Int64 != *(tt.newParentID) {
					t.Errorf("expected parent ID %d, got %v", tt.newParentID, folder.ParentID)
				}
			}
		})
	}
}

//func TestFolderRepository_CascadeDelete(t *testing.T) {
//	folderRepo, userRepo := setupFolderRepositoryTest(t)
//	ctx := testutil.TestContext(t)
//
//	// Setup: Create test user and folder hierarchy
//	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
//	parentID, _ := folderRepo.Insert(ctx, userID, nil, "parent", "/bob/parent")
//	childID, _ := folderRepo.Insert(ctx, userID, parentID, "child", "/bob/parent/child")
//
//	// Delete parent folder
//	err := folderRepo.Delete(ctx, parentID)
//	if err != nil {
//		t.Fatalf("failed to delete parent folder: %v", err)
//	}
//
//	// Verify child folder is also deleted (CASCADE)
//	_, err = folderRepo.GetByID(ctx, childID)
//	if err == nil {
//		t.Error("expected child folder to be deleted via cascade")
//	}
//}
