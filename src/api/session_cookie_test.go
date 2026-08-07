package api

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/src/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionCookie_PlainHTTP(t *testing.T) {
	session := domain.Session{
		SessionKey: "test-key",
		ExpiresAt:  time.Now().Add(time.Hour),
	}
	req := httptest.NewRequest(http.MethodPost, "/api/members/login", nil)

	cookie := sessionCookie(req, &session)

	assert.Equal(t, "session_key", cookie.Name)
	assert.Equal(t, session.SessionKey, cookie.Value)
	assert.Equal(t, "/", cookie.Path)
	assert.True(t, cookie.HttpOnly)
	assert.False(t, cookie.Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.Equal(t, session.ExpiresAt, cookie.Expires)
}

func TestSessionCookie_HTTPS(t *testing.T) {
	session := domain.Session{
		SessionKey: "test-key",
		ExpiresAt:  time.Now().Add(time.Hour),
	}
	req := httptest.NewRequest(http.MethodPost, "/api/members/login", nil)
	req.TLS = &tls.ConnectionState{}

	cookie := sessionCookie(req, &session)

	assert.True(t, cookie.Secure)
}

func TestSessionCookie_ForwardedHTTPS(t *testing.T) {
	session := domain.Session{
		SessionKey: "test-key",
		ExpiresAt:  time.Now().Add(time.Hour),
	}
	req := httptest.NewRequest(http.MethodPost, "/api/members/login", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	cookie := sessionCookie(req, &session)

	assert.True(t, cookie.Secure)
}

func TestSessionCookie_ForwardedHTTP(t *testing.T) {
	session := domain.Session{
		SessionKey: "test-key",
		ExpiresAt:  time.Now().Add(time.Hour),
	}
	req := httptest.NewRequest(http.MethodPost, "/api/members/login", nil)
	req.Header.Set("X-Forwarded-Proto", "http")

	cookie := sessionCookie(req, &session)

	assert.False(t, cookie.Secure)
}

func TestClearSessionCookie_PlainHTTP(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/members/logout", nil)

	cookie := clearSessionCookie(req)

	assert.Equal(t, "session_key", cookie.Name)
	assert.Equal(t, "", cookie.Value)
	assert.Equal(t, -1, cookie.MaxAge)
	assert.False(t, cookie.Secure)
	require.NotNil(t, cookie)
}
