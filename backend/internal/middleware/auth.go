package middleware

import (
	"context"
	"github.com/NiClassic/go-cloud/internal/service"
	"net/http"
)

const ContextUserKey = "user"

type SessionValidator struct {
	svc *service.AuthService
}

func NewSessionValidator(svc *service.AuthService) *SessionValidator {
	return &SessionValidator{svc: svc}
}

func (s *SessionValidator) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := s.svc.GetUserBySessionToken(context.TODO(), cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), ContextUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
