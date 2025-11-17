package service

import (
	"database/sql"
	"github.com/NiClassic/go-cloud/internal/model"
	"github.com/NiClassic/go-cloud/internal/repository"
	"github.com/NiClassic/go-cloud/internal/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFileShareService_CreateFileShare(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := testutil.TestContext(t)

	userRepo := repository.NewUserRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	fileRepo := repository.NewPersonalFileRepository(db)
	shareRepo := repository.NewFileShareRepositoryImpl(db)
	svc := NewFileShareService(*fileRepo, shareRepo)

	ownerID := testutil.InsertTestUser(t, "bob", userRepo)
	recipientID := testutil.InsertTestUser(t, "alice", userRepo)
	otherRecipientID := testutil.InsertTestUser(t, "candice", userRepo)
	folderID := testutil.InsertTestFolder(t, ownerID, "documents", folderRepo)
	fileID := testutil.InsertTestFile(t, ownerID, folderID, "somestuff.txt", fileRepo)

	// Expired time for table
	expiredTime := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name        string
		ownerID     int64
		fileID      int64
		recipientID int64
		perm        string
		expires     *time.Time
		wantErr     error
		wantCount   int
	}{
		{
			name:        "valid share with read permission",
			ownerID:     ownerID,
			fileID:      fileID,
			recipientID: recipientID,
			perm:        "read",
			expires:     nil,
			wantErr:     nil,
			wantCount:   1,
		},
		{
			name:        "valid share with write permission with expiry",
			ownerID:     ownerID,
			fileID:      fileID,
			recipientID: otherRecipientID,
			perm:        "write",
			expires:     &expiredTime,
			wantErr:     nil,
			wantCount:   1,
		},
		{
			name:        "invalid permission",
			ownerID:     ownerID,
			fileID:      fileID,
			recipientID: recipientID,
			perm:        "execute",
			expires:     nil,
			wantErr:     ErrInvalidPerm,
			wantCount:   1, // because prior test inserted one
		},
		{
			name:        "not owner",
			ownerID:     otherRecipientID,
			fileID:      fileID,
			recipientID: recipientID,
			perm:        "read",
			expires:     nil,
			wantErr:     ErrNotOwner,
			wantCount:   1, // because prior tests inserted one
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateFileShare(ctx, tt.ownerID, tt.fileID, tt.recipientID, tt.perm, tt.expires)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			shares, _ := shareRepo.GetByRecipient(ctx, tt.recipientID)
			assert.Len(t, shares, tt.wantCount)
		})
	}

	t.Run("duplicate share second call fails", func(t *testing.T) {
		_, err := svc.CreateFileShare(ctx, ownerID, fileID, recipientID, "read", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "UNIQUE")
	})
}

func TestFileShareService_GetByRecipient(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := testutil.TestContext(t)

	userRepo := repository.NewUserRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	fileRepo := repository.NewPersonalFileRepository(db)
	shareRepo := repository.NewFileShareRepositoryImpl(db)
	svc := NewFileShareService(*fileRepo, shareRepo)

	ownerID := testutil.InsertTestUser(t, "bob", userRepo)
	recipientID := testutil.InsertTestUser(t, "alice", userRepo)
	otherRecipientID := testutil.InsertTestUser(t, "candice", userRepo)
	folderID := testutil.InsertTestFolder(t, ownerID, "documents", folderRepo)
	fileID := testutil.InsertTestFile(t, ownerID, folderID, "somestuff.txt", fileRepo)

	share := &model.FileShare{
		FileID:       fileID,
		SharedWithID: recipientID,
		Permission:   "read",
		ExpiresAt:    sql.NullTime{Valid: false},
	}
	err := shareRepo.Create(ctx, share)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		recipientID int64
		wantErr     error
		wantCount   int
	}{
		{
			name:        "has one shared file",
			recipientID: recipientID,
			wantErr:     nil,
			wantCount:   1,
		},
		{
			name:        "has no shared file",
			recipientID: otherRecipientID,
			wantErr:     nil,
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shares, err := svc.GetByRecipient(ctx, tt.recipientID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			for _, share := range shares {
				assert.Equal(t, share.FileID, fileID)
				assert.Equal(t, share.SharedWithID, recipientID)
			}

			assert.Len(t, shares, tt.wantCount)
		})
	}
}

func TestFileShareService_GetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := testutil.TestContext(t)

	userRepo := repository.NewUserRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	fileRepo := repository.NewPersonalFileRepository(db)
	shareRepo := repository.NewFileShareRepositoryImpl(db)
	svc := NewFileShareService(*fileRepo, shareRepo)

	ownerID := testutil.InsertTestUser(t, "bob", userRepo)
	recipientID := testutil.InsertTestUser(t, "alice", userRepo)
	otherRecipientID := testutil.InsertTestUser(t, "candice", userRepo)
	folderID := testutil.InsertTestFolder(t, ownerID, "documents", folderRepo)
	fileID := testutil.InsertTestFile(t, ownerID, folderID, "somestuff.txt", fileRepo)

	share := &model.FileShare{
		FileID:       fileID,
		SharedWithID: recipientID,
		Permission:   "read",
		ExpiresAt:    sql.NullTime{Valid: false},
	}
	err := shareRepo.Create(ctx, share)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		requester int64
		shareID   int64
		wantErr   error
	}{
		{"non-owner cannot get info", otherRecipientID, share.ID, ErrNotOwner},
		{"owner can get info", ownerID, share.ID, nil},
		{"cannot get invalid share", ownerID, 99999, sql.ErrNoRows},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := svc.GetByID(ctx, tt.requester, tt.shareID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, share.ID, s.ID)
			}
		})
	}
}

func TestFileShareService_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := testutil.TestContext(t)

	userRepo := repository.NewUserRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	fileRepo := repository.NewPersonalFileRepository(db)
	shareRepo := repository.NewFileShareRepositoryImpl(db)
	svc := NewFileShareService(*fileRepo, shareRepo)

	ownerID := testutil.InsertTestUser(t, "bob", userRepo)
	recipientID := testutil.InsertTestUser(t, "alice", userRepo)
	otherRecipientID := testutil.InsertTestUser(t, "candice", userRepo)
	folderID := testutil.InsertTestFolder(t, ownerID, "documents", folderRepo)
	fileID := testutil.InsertTestFile(t, ownerID, folderID, "somestuff.txt", fileRepo)

	share := &model.FileShare{
		FileID:       fileID,
		SharedWithID: recipientID,
		Permission:   "read",
		ExpiresAt:    sql.NullTime{Valid: false},
	}
	err := shareRepo.Create(ctx, share)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		requester int64
		shareID   int64
		wantErr   error
	}{
		{"non-owner cannot delete", otherRecipientID, share.ID, ErrNotOwner},
		{"owner can delete", ownerID, share.ID, nil},
		{"delete non-existent share", ownerID, 99999, sql.ErrNoRows},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteFileShare(ctx, tt.requester, tt.shareID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				shares, err := shareRepo.GetByRecipient(ctx, recipientID)
				assert.NoError(t, err)
				assert.Len(t, shares, 0)
			}
		})
	}
}
