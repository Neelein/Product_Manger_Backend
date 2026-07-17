package api

import (
	"context"
	"errors"
	"net/http"

	"backend/src/domain"
)

type memberCtxKey struct{}

var memberKey = &memberCtxKey{}

func MemberFromContext(ctx context.Context) *domain.Member {
	v := ctx.Value(memberKey)
	if v == nil {
		return nil
	}
	return v.(*domain.Member)
}

func ContextWithMember(ctx context.Context, member *domain.Member) context.Context {
	return context.WithValue(ctx, memberKey, member)
}

func AuthMiddleware(sessionRepo domain.SessionRepository, memberRepo domain.MemberRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_key")
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			newSession, err := sessionRepo.Rotate(r.Context(), cookie.Value)
			if err != nil {
				if errors.Is(err, domain.ErrSessionExpired) {
					writeError(w, http.StatusUnauthorized, "session expired")
					return
				}
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			member, err := memberRepo.GetByID(r.Context(), newSession.MemberID)
			if err != nil || member == nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session_key",
				Value:    newSession.SessionKey,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Expires:  newSession.ExpiresAt,
			})

			ctx := context.WithValue(r.Context(), memberKey, member)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
