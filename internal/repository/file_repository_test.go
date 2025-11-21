package repository_test

import (
	"testing"

	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/testutil"
)

func setupPersonalFileRepositoryTest(t *testing.T) (*repository.FileRepository, *repository.UserRepository, *repository.FolderRepository) {
	t.Helper()

	db := testutil.SetupTestDB(t)

	fileRepo := repository.NewFileRepository(db)
	userRepo := repository.NewUserRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	return fileRepo, userRepo, folderRepo
}

func TestPersonalFileRepository_Insert(t *testing.T) {
	fileRepo, userRepo, folderRepo := setupPersonalFileRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, err := userRepo.Insert(ctx, "bob", "hashedpass")
	if err != nil {
		t.Fatal("could not insert test user")
	}

	folderID, err := folderRepo.Insert(ctx, userID, nil, "bob", "bob")
	if err != nil {
		t.Fatal("could not insert test folder")
	}

	tests := []struct {
		name     string
		filename string
		mimeType string
		location string
		hash     string
		userID   int64
		size     int64
		folderID int64
		wantErr  bool
	}{
		{
			name:     "valid file",
			filename: "test.txt",
			mimeType: "text/plain",
			location: "bob/test.txt",
			hash:     "abc123",
			userID:   userID,
			size:     1024,
			folderID: folderID,
			wantErr:  false,
		},
		{
			name:     "another valid file",
			filename: "image.png",
			mimeType: "image/png",
			location: "bob/image.png",
			hash:     "def456",
			userID:   userID,
			size:     2048,
			folderID: folderID,
			wantErr:  false,
		},
		{
			name:     "invalid user ID",
			filename: "test.txt",
			mimeType: "text/plain",
			location: "bob/test.txt",
			hash:     "ghi789",
			userID:   99999,
			size:     512,
			folderID: folderID,
			wantErr:  true,
		},
		{
			name:     "invalid folder ID",
			filename: "test.txt",
			mimeType: "text/plain",
			location: "bob/test.txt",
			hash:     "jkl012",
			userID:   userID,
			size:     512,
			folderID: 99999,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileID, err := fileRepo.Insert(ctx, tt.filename, tt.mimeType, tt.location, tt.hash, tt.userID, tt.size, tt.folderID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if fileID <= 0 {
				t.Error("expected valid file ID")
			}
		})
	}
}

func TestPersonalFileRepository_GetByUser(t *testing.T) {
	fileRepo, userRepo, folderRepo := setupPersonalFileRepositoryTest(t)
	ctx := testutil.TestContext(t)

	user1ID, _ := userRepo.Insert(ctx, "user1", "pass1")
	user2ID, _ := userRepo.Insert(ctx, "user2", "pass2")

	folder1ID, _ := folderRepo.Insert(ctx, user1ID, nil, "user1", "user1")
	folder2ID, _ := folderRepo.Insert(ctx, user2ID, nil, "user2", "user2")

	fileRepo.Insert(ctx, "file1.txt", "text/plain", "user1/file1.txt", "hash1", user1ID, 100, folder1ID)
	fileRepo.Insert(ctx, "file2.txt", "text/plain", "user1/file2.txt", "hash2", user1ID, 200, folder1ID)

	fileRepo.Insert(ctx, "file3.txt", "text/plain", "user2/file3.txt", "hash3", user2ID, 300, folder2ID)

	tests := []struct {
		name          string
		userID        int64
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "user1 files",
			userID:        user1ID,
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:          "user2 files",
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
			files, err := fileRepo.GetByUser(ctx, tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(files) != tt.expectedCount {
				t.Errorf("expected %d files, got %d", tt.expectedCount, len(files))
			}

			for _, file := range files {
				if file.UserID != tt.userID {
					t.Errorf("file %d belongs to user %d, expected %d", file.ID, file.UserID, tt.userID)
				}
			}
		})
	}
}

func TestPersonalFileRepository_GetByUserAndFolder(t *testing.T) {
	fileRepo, userRepo, folderRepo := setupPersonalFileRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
	folder1ID, _ := folderRepo.Insert(ctx, userID, nil, "folder1", "bob/folder1")
	folder2ID, _ := folderRepo.Insert(ctx, userID, nil, "folder2", "bob/folder2")

	fileRepo.Insert(ctx, "file1.txt", "text/plain", "bob/folder1/file1.txt", "hash1", userID, 100, folder1ID)
	fileRepo.Insert(ctx, "file2.txt", "text/plain", "bob/folder1/file2.txt", "hash2", userID, 200, folder1ID)
	fileRepo.Insert(ctx, "file3.txt", "text/plain", "bob/folder2/file3.txt", "hash3", userID, 300, folder2ID)

	tests := []struct {
		name          string
		userID        int64
		folderID      int64
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "folder1 files",
			userID:        userID,
			folderID:      folder1ID,
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:          "folder2 files",
			userID:        userID,
			folderID:      folder2ID,
			expectedCount: 1,
			wantErr:       false,
		},
		{
			name:          "wrong user ID",
			userID:        99999,
			folderID:      folder1ID,
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name:          "non-existent folder",
			userID:        userID,
			folderID:      99999,
			expectedCount: 0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := fileRepo.GetByUserAndFolder(ctx, tt.userID, tt.folderID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(files) != tt.expectedCount {
				t.Errorf("expected %d files, got %d", tt.expectedCount, len(files))
			}

			for _, file := range files {
				if file.UserID != tt.userID {
					t.Errorf("file belongs to user %d, expected %d", file.UserID, tt.userID)
				}

				if !file.FolderID.Valid || file.FolderID.Int64 != tt.folderID {
					t.Errorf("file belongs to folder %v, expected %d", file.FolderID, tt.folderID)
				}
			}
		})
	}
}

func TestPersonalFileRepository_GetById(t *testing.T) {
	fileRepo, userRepo, folderRepo := setupPersonalFileRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
	folderID, _ := folderRepo.Insert(ctx, userID, nil, "bob", "bob")

	fileName := "test.txt"
	mimeType := "text/plain"
	location := "bob/test.txt"
	hash := "abc123"
	size := int64(1024)

	fileID, err := fileRepo.Insert(ctx, fileName, mimeType, location, hash, userID, size, folderID)
	if err != nil {
		t.Fatal("could not insert test file")
	}

	tests := []struct {
		name    string
		fileID  int64
		wantErr bool
	}{
		{
			name:    "existing file",
			fileID:  fileID,
			wantErr: false,
		},
		{
			name:    "non-existent file",
			fileID:  99999,
			wantErr: true,
		},
		{
			name:    "zero file ID",
			fileID:  0,
			wantErr: true,
		},
		{
			name:    "negative file ID",
			fileID:  -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := fileRepo.GetById(ctx, tt.fileID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if file.ID != fileID {
				t.Errorf("expected file ID %d, got %d", fileID, file.ID)
			}

			if file.Name != fileName {
				t.Errorf("expected name %s, got %s", fileName, file.Name)
			}

			if file.MimeType != mimeType {
				t.Errorf("expected mime type %s, got %s", mimeType, file.MimeType)
			}

			if file.Location != location {
				t.Errorf("expected location %s, got %s", location, file.Location)
			}

			if file.Hash != hash {
				t.Errorf("expected hash %s, got %s", hash, file.Hash)
			}

			if file.Size != size {
				t.Errorf("expected size %d, got %d", size, file.Size)
			}

			if file.UserID != userID {
				t.Errorf("expected user ID %d, got %d", userID, file.UserID)
			}

			if !file.FolderID.Valid || file.FolderID.Int64 != folderID {
				t.Errorf("expected folder ID %d, got %v", folderID, file.FolderID)
			}

			if file.CreatedAt.IsZero() {
				t.Error("expected non-zero created_at timestamp")
			}
		})
	}
}

func TestPersonalFileRepository_UpdateFolder(t *testing.T) {
	fileRepo, userRepo, folderRepo := setupPersonalFileRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
	folder1ID, _ := folderRepo.Insert(ctx, userID, nil, "folder1", "bob/folder1")
	folder2ID, _ := folderRepo.Insert(ctx, userID, nil, "folder2", "bob/folder2")

	fileID, err := fileRepo.Insert(ctx, "test.txt", "text/plain", "bob/folder1/test.txt", "hash", userID, 100, folder1ID)
	if err != nil {
		t.Fatal("could not insert test file")
	}

	tests := []struct {
		name        string
		fileID      int64
		newFolderID int64
		wantErr     bool
	}{
		{
			name:        "move to different folder",
			fileID:      fileID,
			newFolderID: folder2ID,
			wantErr:     false,
		},
		// Update with zero targets is not an error
		{
			name:        "update non-existent file",
			fileID:      99999,
			newFolderID: folder1ID,
			wantErr:     false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fileRepo.UpdateFolder(ctx, tt.fileID, tt.newFolderID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.fileID == fileID {
				file, err := fileRepo.GetById(ctx, tt.fileID)
				if err != nil {
					t.Fatalf("could not get updated file: %v", err)
				}

				if !file.FolderID.Valid || file.FolderID.Int64 != tt.newFolderID {
					t.Errorf("expected folder ID %d, got %v", tt.newFolderID, file.FolderID)
				}
			}
		})
	}
}

func TestPersonalFileRepository_Delete(t *testing.T) {
	fileRepo, userRepo, folderRepo := setupPersonalFileRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
	folderID, _ := folderRepo.Insert(ctx, userID, nil, "bob", "bob")

	fileID, err := fileRepo.Insert(ctx, "todelete.txt", "text/plain", "bob/todelete.txt", "hash", userID, 100, folderID)
	if err != nil {
		t.Fatal("could not insert test file")
	}

	tests := []struct {
		name    string
		fileID  int64
		wantErr bool
	}{
		{
			name:    "delete existing file",
			fileID:  fileID,
			wantErr: false,
		},
		// Delete with no matches is not an error
		{
			name:    "delete non-existent file",
			fileID:  99999,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fileRepo.Delete(ctx, tt.fileID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.fileID == fileID {
				_, err := fileRepo.GetById(ctx, tt.fileID)
				if err == nil {
					t.Error("expected error when getting deleted file")
				}
			}
		})
	}
}

func TestPersonalFileRepository_FolderSetNull(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := testutil.TestContext(t)

	fileRepo := repository.NewFileRepository(db)
	userRepo := repository.NewUserRepository(db)
	folderRepo := repository.NewFolderRepository(db)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
	folderID, _ := folderRepo.Insert(ctx, userID, nil, "bob", "bob")
	fileID, _ := fileRepo.Insert(ctx, "test.txt", "text/plain", "bob/test.txt", "hash", userID, 100, folderID)

	err := folderRepo.Delete(ctx, folderID)
	if err != nil {
		t.Fatalf("could not delete folder: %v", err)
	}

	file, err := fileRepo.GetById(ctx, fileID)
	if err != nil {
		t.Fatalf("could not get file: %v", err)
	}

	if file.FolderID.Valid {
		t.Errorf("expected folder_id to be 0 (NULL), got %v", file.FolderID)
	}
}

func TestPersonalFileRepository_MultipleFilesInFolder(t *testing.T) {
	fileRepo, userRepo, folderRepo := setupPersonalFileRepositoryTest(t)
	ctx := testutil.TestContext(t)

	userID, _ := userRepo.Insert(ctx, "bob", "hashedpass")
	folderID, _ := folderRepo.Insert(ctx, userID, nil, "bob", "bob")

	fileNames := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, name := range fileNames {
		_, err := fileRepo.Insert(ctx, name, "text/plain", "bob/"+name, "hash"+name, userID, 100, folderID)
		if err != nil {
			t.Fatalf("could not insert file %s: %v", name, err)
		}
	}

	files, err := fileRepo.GetByUserAndFolder(ctx, userID, folderID)
	if err != nil {
		t.Fatalf("could not get files: %v", err)
	}

	if len(files) != len(fileNames) {
		t.Errorf("expected %d files, got %d", len(fileNames), len(files))
	}

	foundNames := make(map[string]bool)
	for _, file := range files {
		foundNames[file.Name] = true
	}

	for _, name := range fileNames {
		if !foundNames[name] {
			t.Errorf("file %s not found", name)
		}
	}
}
