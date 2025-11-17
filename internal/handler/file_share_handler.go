package handler

import (
	"github.com/NiClassic/go-cloud/config"
	"github.com/NiClassic/go-cloud/internal/logger"
	"github.com/NiClassic/go-cloud/internal/service"
	"net/http"
)

type FileShareHandler struct {
	*baseHandler
	svc *service.FileShareService
}

func NewFileShareHandler(cfg *config.Config, r *Renderer, svc *service.FileShareService) *FileShareHandler {
	return &FileShareHandler{newBaseHandler(cfg, r), svc}
}

func (h *FileShareHandler) GetSharedWithUser(w http.ResponseWriter, r *http.Request) {
	logger.Request(r)
	switch r.Method {
	case "GET":
		user := ExtractUserOrRedirect(w, r)
		if user == nil {
			return
		}
		sharedFiles, err := h.svc.GetSharedFiles(r.Context(), user.ID)
		if err != nil {
			logger.Error("could not retrieve file shares: %v", err)
		}
		h.r.Render(w, true, SharePage, "Your Shares", map[string]any{
			"Files": sharedFiles,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.InvalidMethod(r)
	}
}
