package middleware

import (
	"context"
	"github.com/NiClassic/go-cloud/internal/service"
	"net/http"
)

type GuestOnly struct {
	svc *service.AuthService
}

func NewGuestOnly(svc *service.AuthService) *GuestOnly {
	return &GuestOnly{svc: svc}
}

func (g *GuestOnly) WithoutAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err == nil {
			user, err := g.svc.GetUserBySessionToken(context.TODO(), cookie.Value)
			if err == nil && user != nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
