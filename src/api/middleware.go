package api

import (
	"context"
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

			newSession, err := sessionRepo.GetByKey(r.Context(), cookie.Value)
			if err != nil || newSession == nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			member, err := memberRepo.GetByID(r.Context(), newSession.MemberID)
			if err != nil || member == nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			http.SetCookie(w, sessionCookie(r, newSession))

			ctx := context.WithValue(r.Context(), memberKey, member)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole wraps an authenticated middleware chain: it first runs the auth
// middleware to populate the member context, then rejects the request unless the
// authenticated member has the required role.
func RequireRole(role string, auth func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			member := MemberFromContext(r.Context())
			if member == nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			if member.Role != role {
				writeError(w, http.StatusForbidden, "forbidden")
				return
			}
			handler.ServeHTTP(w, r)
		}))
	}
}
