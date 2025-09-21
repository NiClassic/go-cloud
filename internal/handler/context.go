package handler

import (
	"github.com/NiClassic/go-cloud/internal/logger"
	"github.com/NiClassic/go-cloud/internal/middleware"
	"github.com/NiClassic/go-cloud/internal/model"
	"net/http"
)

func ExtractUserOrRedirect(w http.ResponseWriter, r *http.Request) *model.User {
	user, ok := r.Context().Value(middleware.UserKey).(*model.User)
	if !ok || user == nil {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		logger.Info("could not extract user from context")
		return nil
	}
	return user
}
