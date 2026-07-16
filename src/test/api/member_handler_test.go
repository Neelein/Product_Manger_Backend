//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/src/api"
	"backend/src/database"
	"backend/src/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMemberHandler() (*database.MemberRepositoryPGX, *database.SessionRepositoryPGX, *api.MemberHandler) {
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	sessionRepo := database.NewSessionRepositoryPGX(testPool)
	handler := api.NewMemberHandler(memberRepo, sessionRepo)
	return memberRepo, sessionRepo, handler
}

func cleanupMembers(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE members CASCADE")
	require.NoError(t, err)
	_, err = testPool.Exec(context.Background(), "TRUNCATE TABLE sessions CASCADE")
	require.NoError(t, err)
}

func TestHandler_Register(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	body, _ := json.Marshal(domain.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "John Doe",
	})

	w := executeRequest(http.MethodPost, "/api/members/register", body, handler.RegisterMember)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp domain.MemberResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "user@example.com", resp.Email)
	assert.Equal(t, "John Doe", resp.Name)
}

func TestHandler_Register_DuplicateEmail(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	body, _ := json.Marshal(domain.RegisterRequest{
		Email:    "dup@example.com",
		Password: "password123",
		Name:     "User",
	})

	w := executeRequest(http.MethodPost, "/api/members/register", body, handler.RegisterMember)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = executeRequest(http.MethodPost, "/api/members/register", body, handler.RegisterMember)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandler_Login(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	regBody, _ := json.Marshal(domain.RegisterRequest{
		Email:    "login@example.com",
		Password: "mypassword",
		Name:     "Login User",
	})
	w := executeRequest(http.MethodPost, "/api/members/register", regBody, handler.RegisterMember)
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody, _ := json.Marshal(domain.LoginRequest{
		Email:    "login@example.com",
		Password: "mypassword",
	})
	w = executeRequest(http.MethodPost, "/api/members/login", loginBody, handler.LoginMember)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.LoginResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "login@example.com", resp.Member.Email)
	assert.Equal(t, "Login User", resp.Member.Name)

	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_key" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)
	assert.NotEmpty(t, sessionCookie.Value)
	assert.True(t, sessionCookie.HttpOnly)
}

func TestHandler_Login_WrongPassword(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	regBody, _ := json.Marshal(domain.RegisterRequest{
		Email:    "wrong@example.com",
		Password: "correctpw",
		Name:     "User",
	})
	w := executeRequest(http.MethodPost, "/api/members/register", regBody, handler.RegisterMember)
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody, _ := json.Marshal(domain.LoginRequest{
		Email:    "wrong@example.com",
		Password: "wrongpw",
	})
	w = executeRequest(http.MethodPost, "/api/members/login", loginBody, handler.LoginMember)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_Me(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	regBody, _ := json.Marshal(domain.RegisterRequest{
		Email:    "me@example.com",
		Password: "password",
		Name:     "Me User",
	})
	w := executeRequest(http.MethodPost, "/api/members/register", regBody, handler.RegisterMember)
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody, _ := json.Marshal(domain.LoginRequest{
		Email:    "me@example.com",
		Password: "password",
	})
	w = executeRequest(http.MethodPost, "/api/members/login", loginBody, handler.LoginMember)
	require.Equal(t, http.StatusOK, w.Code)

	sessionCookie := w.Result().Cookies()[0]

	req := httptest.NewRequest(http.MethodGet, "/api/members/me", nil)
	req.AddCookie(sessionCookie)
	w = httptest.NewRecorder()
	handler.GetCurrentMember(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.MemberResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "me@example.com", resp.Email)
	assert.Equal(t, "Me User", resp.Name)
}

func TestHandler_Me_NoCookie(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	w := executeRequest(http.MethodGet, "/api/members/me", nil, handler.GetCurrentMember)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_Me_ExpiredSession(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	regBody, _ := json.Marshal(domain.RegisterRequest{
		Email:    "expired@example.com",
		Password: "password",
		Name:     "Expired User",
	})
	w := executeRequest(http.MethodPost, "/api/members/register", regBody, handler.RegisterMember)
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody, _ := json.Marshal(domain.LoginRequest{
		Email:    "expired@example.com",
		Password: "password",
	})
	w = executeRequest(http.MethodPost, "/api/members/login", loginBody, handler.LoginMember)
	require.Equal(t, http.StatusOK, w.Code)

	sessionCookie := w.Result().Cookies()[0]

	_, err := testPool.Exec(context.Background(),
		`UPDATE sessions SET expires_at = $1 WHERE session_key = $2`,
		time.Now().Add(-1*time.Hour), sessionCookie.Value)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/members/me", nil)
	req.AddCookie(sessionCookie)
	w = httptest.NewRecorder()
	handler.GetCurrentMember(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_UpdateMember(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	regBody, _ := json.Marshal(domain.RegisterRequest{
		Email:    "update@example.com",
		Password: "password",
		Name:     "Original Name",
	})
	w := executeRequest(http.MethodPost, "/api/members/register", regBody, handler.RegisterMember)
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody, _ := json.Marshal(domain.LoginRequest{
		Email:    "update@example.com",
		Password: "password",
	})
	w = executeRequest(http.MethodPost, "/api/members/login", loginBody, handler.LoginMember)
	require.Equal(t, http.StatusOK, w.Code)

	sessionCookie := w.Result().Cookies()[0]

	t.Run("successful update", func(t *testing.T) {
		updateBody, _ := json.Marshal(domain.UpdateMemberRequest{
			Email: "updated@example.com",
			Name:  "Updated Name",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/members/update", bytes.NewReader(updateBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)
		w = httptest.NewRecorder()
		handler.UpdateMember(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.MemberResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, "updated@example.com", resp.Email)
		assert.Equal(t, "Updated Name", resp.Name)
	})

	t.Run("update without auth cookie", func(t *testing.T) {
		updateBody, _ := json.Marshal(domain.UpdateMemberRequest{
			Email: "noauth@example.com",
			Name:  "No Auth",
		})
		w := executeRequest(http.MethodPost, "/api/members/update", updateBody, handler.UpdateMember)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("update with empty fields", func(t *testing.T) {
		updateBody, _ := json.Marshal(domain.UpdateMemberRequest{
			Email: "",
			Name:  "",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/members/update", bytes.NewReader(updateBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)
		w = httptest.NewRecorder()
		handler.UpdateMember(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("update email to existing email", func(t *testing.T) {
		otherRegBody, _ := json.Marshal(domain.RegisterRequest{
			Email:    "existing@example.com",
			Password: "password",
			Name:     "Existing User",
		})
		w := executeRequest(http.MethodPost, "/api/members/register", otherRegBody, handler.RegisterMember)
		require.Equal(t, http.StatusCreated, w.Code)

		updateBody, _ := json.Marshal(domain.UpdateMemberRequest{
			Email: "existing@example.com",
			Name:  "Updated Name",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/members/update", bytes.NewReader(updateBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)
		w = httptest.NewRecorder()
		handler.UpdateMember(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestHandler_Logout(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	regBody, _ := json.Marshal(domain.RegisterRequest{
		Email:    "logout@example.com",
		Password: "password",
		Name:     "Logout User",
	})
	w := executeRequest(http.MethodPost, "/api/members/register", regBody, handler.RegisterMember)
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody, _ := json.Marshal(domain.LoginRequest{
		Email:    "logout@example.com",
		Password: "password",
	})
	w = executeRequest(http.MethodPost, "/api/members/login", loginBody, handler.LoginMember)
	require.Equal(t, http.StatusOK, w.Code)

	sessionCookie := w.Result().Cookies()[0]

	req := httptest.NewRequest(http.MethodPost, "/api/members/logout", nil)
	req.AddCookie(sessionCookie)
	w = httptest.NewRecorder()
	handler.LogoutMember(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	clearCookie := w.Result().Cookies()[0]
	assert.Equal(t, "session_key", clearCookie.Name)
	assert.Equal(t, "", clearCookie.Value)
	assert.Equal(t, -1, clearCookie.MaxAge)
}
