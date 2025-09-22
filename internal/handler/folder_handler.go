package handler

import (
	"github.com/NiClassic/go-cloud/internal/service"
)

type FolderHandler struct {
	folderSvc *service.FolderService
	fileSvc   *service.PersonalFileService
}

func NewFolderHandler(folderSvc *service.FolderService, fileSvc *service.PersonalFileService) *FolderHandler {
	return &FolderHandler{
		folderSvc: folderSvc,
		fileSvc:   fileSvc,
	}
}
