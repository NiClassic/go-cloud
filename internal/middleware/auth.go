package middleware

import (
	"context"
	"github.com/NiClassic/go-cloud/internal/logger"
	"net/http"

	"github.com/NiClassic/go-cloud/internal/service"
)

const (
	cookieName   = "session_token"
	redirectPath = "/login"
)

type ctxKey int

const (
	UserKey ctxKey = iota
)

type SessionValidator struct{ svc *service.AuthService }

func NewSessionValidator(svc *service.AuthService) *SessionValidator {
	return &SessionValidator{svc: svc}
}

func (s *SessionValidator) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			logger.Error("could not find cookie: %v", err)
			http.Redirect(w, r, redirectPath, http.StatusSeeOther)
			return
		}

		user, err := s.svc.GetUserBySessionToken(r.Context(), cookie.Value)
		if err != nil {
			logger.Error("could not get user: %v", err)
			http.Redirect(w, r, redirectPath, http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type GuestOnly struct{ svc *service.AuthService }

func NewGuestOnly(svc *service.AuthService) *GuestOnly {
	return &GuestOnly{svc: svc}
}

func (g *GuestOnly) WithoutAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(cookieName); err == nil {
			if user, err := g.svc.GetUserBySessionToken(r.Context(), cookie.Value); err == nil && user != nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
