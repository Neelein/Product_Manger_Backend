package http

import (
	"net/http"
	"strings"
)

func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func sessionCookie(r *http.Request, s *Session) *http.Cookie {
	return &http.Cookie{Name: "session_key", Value: s.SessionKey, Path: "/", HttpOnly: true, Secure: isHTTPS(r), SameSite: http.SameSiteLaxMode, Expires: s.ExpiresAt}
}

func clearSessionCookie(r *http.Request) *http.Cookie {
	return &http.Cookie{Name: "session_key", Value: "", Path: "/", HttpOnly: true, Secure: isHTTPS(r), SameSite: http.SameSiteLaxMode, MaxAge: -1}
}

func SessionCookieForCompatibility(r *http.Request, s *Session) *http.Cookie {
	return sessionCookie(r, s)
}
func ClearSessionCookieForCompatibility(r *http.Request) *http.Cookie { return clearSessionCookie(r) }
