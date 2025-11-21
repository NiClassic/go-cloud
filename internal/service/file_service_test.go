package service_test

import (
	"bytes"
	"fmt"
	"github.com/NiClassic/go-cloud/internal/path"
	"io"
	"mime/multipart"
	"net/textproto"
	"testing"

	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/service"
	"github.com/NiClassic/go-cloud/internal/storage"
	"github.com/NiClassic/go-cloud/internal/testutil"
)

func setupPersonalFileTest(t *testing.T) (*service.FileService, *service.FolderService, *model.User, int64) {
	t.Helper()

	db := testutil.SetupTestDB(t)
	ctx := testutil.TestContext(t)
	tmpDir := testutil.SetupTestStorage(t)

	userRepo := repository.NewUserRepository(db)
	fileRepo := repository.NewFileRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	fileShareRepo := repository.NewFileShareRepositoryImpl(db)
	st := storage.NewIOStorage(tmpDir)
	c := path.New(tmpDir)

	fileSvc := service.NewFileService(st, fileRepo, userRepo, fileShareRepo, c)
	folderSvc := service.NewFolderService(folderRepo, fileRepo, st, c)

	// Create a test user
	userID, err := userRepo.Insert(ctx, "testuser", "hashedpass")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	user := &model.User{
		ID:       userID,
		Username: "testuser",
	}

	// Create root folder
	rootFolder, err := folderSvc.CreateFolder(ctx, userID, "testuser", -1, "testuser", "/")
	if err != nil {
		t.Fatalf("failed to create root folder: %v", err)
	}

	return fileSvc, folderSvc, user, rootFolder.ID
}

func createMultipartReader(t *testing.T, files map[string]string) *multipart.Reader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for filename, content := range files {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="files"; filename="`+filename+`"`)
		h.Set("Content-Type", "text/plain")

		part, err := writer.CreatePart(h)
		if err != nil {
			t.Fatalf("failed to create part: %v", err)
		}

		if _, err := io.WriteString(part, content); err != nil {
			t.Fatalf("failed to write content: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	reader := multipart.NewReader(body, writer.Boundary())
	return reader
}

func TestPersonalFileService_StoreFiles(t *testing.T) {
	fileSvc, _, user, folderID := setupPersonalFileTest(t)
	ctx := testutil.TestContext(t)

	tests := []struct {
		name        string
		files       map[string]string
		folderPath  string
		expectCount int
	}{
		{
			name: "single file",
			files: map[string]string{
				"test.txt": "Hello, World!",
			},
			folderPath:  "/",
			expectCount: 1,
		},
		{
			name: "multiple files",
			files: map[string]string{
				"file1.txt": "Content 1",
				"file2.txt": "Content 2",
				"file3.txt": "Content 3",
			},
			folderPath:  "/",
			expectCount: 3,
		},
		{
			name: "file with special characters",
			files: map[string]string{
				"test-file_123.txt": "Special content",
			},
			folderPath:  "/",
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := createMultipartReader(t, tt.files)

			err := fileSvc.StoreFiles(ctx, user, reader, folderID, tt.folderPath)
			if err != nil {
				t.Fatalf("failed to store files: %v", err)
			}

			files, err := fileSvc.GetUserFiles(ctx, user)
			if err != nil {
				t.Fatalf("failed to get user files: %v", err)
			}

			if len(files) < tt.expectCount {
				t.Errorf("expected at least %d files, got %d", tt.expectCount, len(files))
			}
		})
	}
}

func TestPersonalFileService_GetUserFiles(t *testing.T) {
	fileSvc, _, user, folderID := setupPersonalFileTest(t)
	ctx := testutil.TestContext(t)

	// Store some test files
	files := map[string]string{
		"doc1.txt": "Document 1",
		"doc2.txt": "Document 2",
	}
	reader := createMultipartReader(t, files)

	err := fileSvc.StoreFiles(ctx, user, reader, folderID, "/")
	if err != nil {
		t.Fatalf("failed to store files: %v", err)
	}

	t.Run("get all user files", func(t *testing.T) {
		userFiles, err := fileSvc.GetUserFiles(ctx, user)
		if err != nil {
			t.Fatalf("failed to get user files: %v", err)
		}

		if len(userFiles) != len(files) {
			t.Errorf("expected %d files, got %d", len(files), len(userFiles))
		}

		// Verify files belong to the user
		for _, file := range userFiles {
			if file.UserID != user.ID {
				t.Errorf("expected user ID %d, got %d", user.ID, file.UserID)
			}

			if file.Name == "" {
				t.Error("file name should not be empty")
			}

			if file.Size == 0 {
				t.Error("file size should not be zero")
			}

			if file.Hash == "" {
				t.Error("file hash should not be empty")
			}
		}
	})
}

func TestPersonalFileService_GetFileById(t *testing.T) {
	fileSvc, _, user, folderID := setupPersonalFileTest(t)
	ctx := testutil.TestContext(t)

	// Store a test file
	files := map[string]string{
		"test.txt": "Test content",
	}
	reader := createMultipartReader(t, files)

	err := fileSvc.StoreFiles(ctx, user, reader, folderID, "/")
	if err != nil {
		t.Fatalf("failed to store file: %v", err)
	}

	// Get the stored file
	userFiles, err := fileSvc.GetUserFiles(ctx, user)
	if err != nil || len(userFiles) == 0 {
		t.Fatalf("failed to get user files: %v", err)
	}

	fileID := userFiles[0].ID

	t.Run("get existing file", func(t *testing.T) {
		file, err := fileSvc.GetFileById(ctx, fileID)
		if err != nil {
			t.Fatalf("failed to get file by ID: %v", err)
		}

		if file.ID != fileID {
			t.Errorf("expected file ID %d, got %d", fileID, file.ID)
		}

		if file.UserID != user.ID {
			t.Errorf("expected user ID %d, got %d", user.ID, file.UserID)
		}
	})

	t.Run("get non-existent file", func(t *testing.T) {
		_, err := fileSvc.GetFileById(ctx, 99999)
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
}

func TestPersonalFileService_DeleteFile(t *testing.T) {
	fileSvc, _, user, folderID := setupPersonalFileTest(t)
	ctx := testutil.TestContext(t)

	// Store a test file
	files := map[string]string{
		"todelete.txt": "Content to delete",
	}
	reader := createMultipartReader(t, files)

	err := fileSvc.StoreFiles(ctx, user, reader, folderID, "/")
	if err != nil {
		t.Fatalf("failed to store file: %v", err)
	}

	userFiles, err := fileSvc.GetUserFiles(ctx, user)
	if err != nil || len(userFiles) == 0 {
		t.Fatalf("failed to get user files: %v", err)
	}

	fileID := userFiles[0].ID

	t.Run("delete existing file", func(t *testing.T) {
		err := fileSvc.DeleteFile(ctx, user, fileID)
		if err != nil {
			t.Fatalf("failed to delete file: %v", err)
		}

		// Verify deletion
		_, err = fileSvc.GetFileById(ctx, fileID)
		if err == nil {
			t.Error("expected error when getting deleted file")
		}
	})

	t.Run("delete non-existent file", func(t *testing.T) {
		err := fileSvc.DeleteFile(ctx, user, 99999)
		if err == nil {
			t.Error("expected error when deleting non-existent file")
		}
	})

	t.Run("unauthorized deletion", func(t *testing.T) {
		// Create another file
		reader := createMultipartReader(t, map[string]string{"another.txt": "content"})
		_ = fileSvc.StoreFiles(ctx, user, reader, folderID, "/")

		files, _ := fileSvc.GetUserFiles(ctx, user)
		if len(files) == 0 {
			t.Skip("no files to test unauthorized deletion")
		}

		// Try to delete with different user
		otherUser := &model.User{ID: user.ID + 999, Username: "otheruser"}
		err := fileSvc.DeleteFile(ctx, otherUser, files[0].ID)
		if err == nil {
			t.Error("expected error when deleting file as wrong user")
		}
	})
}

func TestPersonalFileService_FileHashing(t *testing.T) {
	fileSvc, _, user, folderID := setupPersonalFileTest(t)
	ctx := testutil.TestContext(t)

	content := "This is test content for hashing"
	files := map[string]string{
		"hash1.txt": content,
		"hash2.txt": content, // Same content
	}

	reader := createMultipartReader(t, files)
	err := fileSvc.StoreFiles(ctx, user, reader, folderID, "/")
	if err != nil {
		t.Fatalf("failed to store files: %v", err)
	}

	userFiles, err := fileSvc.GetUserFiles(ctx, user)
	if err != nil {
		t.Fatalf("failed to get user files: %v", err)
	}

	if len(userFiles) < 2 {
		t.Fatal("expected at least 2 files")
	}

	// Find the two files we just uploaded
	var hash1, hash2 string
	for _, f := range userFiles {
		if f.Name == "hash1.txt" {
			hash1 = f.Hash
		} else if f.Name == "hash2.txt" {
			hash2 = f.Hash
		}
	}

	t.Run("identical content produces identical hash", func(t *testing.T) {
		if hash1 == "" || hash2 == "" {
			t.Fatal("failed to find uploaded files")
		}

		if hash1 != hash2 {
			t.Errorf("expected identical hashes for identical content, got %s and %s", hash1, hash2)
		}
	})

	t.Run("hash is not empty", func(t *testing.T) {
		if hash1 == "" {
			t.Error("file hash should not be empty")
		}

		// SHA256 produces 64 character hex string
		if len(hash1) != 64 {
			t.Errorf("expected hash length 64, got %d", len(hash1))
		}
	})
}

func TestPersonalFileService_FileSizeTracking(t *testing.T) {
	fileSvc, _, user, folderID := setupPersonalFileTest(t)
	ctx := testutil.TestContext(t)

	tests := []struct {
		name         string
		content      string
		expectedSize int64
	}{
		{
			name:         "small file",
			content:      "small",
			expectedSize: 5,
		},
		{
			name:         "medium file",
			content:      "This is a medium sized file with some content",
			expectedSize: int64(len("This is a medium sized file with some content")),
		},
		{
			name:         "empty file",
			content:      "",
			expectedSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFileName := fmt.Sprintf("size_test_%s.txt", tt.name)
			files := map[string]string{
				testFileName: tt.content,
			}

			reader := createMultipartReader(t, files)
			err := fileSvc.StoreFiles(ctx, user, reader, folderID, "/")
			if err != nil {
				t.Fatalf("failed to store file: %v", err)
			}

			userFiles, err := fileSvc.GetUserFiles(ctx, user)
			if err != nil {
				t.Fatalf("failed to get user files: %v", err)
			}

			var found bool
			for _, f := range userFiles {
				if f.Name == testFileName {
					if f.Size != tt.expectedSize {
						t.Errorf("expected size %d, got %d", tt.expectedSize, f.Size)
					}
					found = true
					break
				}
			}

			if !found {
				t.Error("uploaded file not found")
			}
		})
	}
}

func TestPersonalFileService_MimeTypeDetection(t *testing.T) {
	fileSvc, _, user, folderID := setupPersonalFileTest(t)
	ctx := testutil.TestContext(t)

	tests := []struct {
		name         string
		filename     string
		content      string
		expectedMime string
	}{
		{
			name:         "text file",
			filename:     "test.txt",
			content:      "plain text",
			expectedMime: "text/plain",
		},
		{
			name:         "json file",
			filename:     "data.json",
			content:      `{"key": "value"}`,
			expectedMime: "application/json",
		},
		{
			name:         "html file",
			filename:     "page.html",
			content:      "<html><body>test</body></html>",
			expectedMime: "text/html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{
				tt.filename: tt.content,
			}

			reader := createMultipartReader(t, files)
			err := fileSvc.StoreFiles(ctx, user, reader, folderID, "/")
			if err != nil {
				t.Fatalf("failed to store file: %v", err)
			}

			userFiles, err := fileSvc.GetUserFiles(ctx, user)
			if err != nil {
				t.Fatalf("failed to get user files: %v", err)
			}

			var found bool
			for _, f := range userFiles {
				if f.Name == tt.filename {
					if f.MimeType == "" {
						t.Error("MIME type should not be empty")
					}
					found = true
					break
				}
			}

			if !found {
				t.Error("uploaded file not found")
			}
		})
	}
}
