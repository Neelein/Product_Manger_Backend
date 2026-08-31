//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	api "backend/src/adapter/http"
	domain "backend/src/adapter/http"
	database "backend/src/adapter/postgres"
	"backend/src/adapter/session"
	"backend/src/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var codeSeq atomic.Uint64

func setupMemberHandler() (*database.MemberRepositoryPGX, *session.SessionCache, *api.MemberHandler) {
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	sessionCache := session.NewSessionCache(time.Hour)
	handler := composeMemberHandler(memberRepo, sessionCache, database.NewRegistrationCodeRepositoryPGX(testPool))
	return memberRepo, sessionCache, handler
}

func cleanupMembers(t *testing.T) {
	t.Helper()
	require.NoError(t, testHarness.Reset(context.Background()))
}

// seedCode creates a fresh unused registration code and returns its raw value.
func seedCode(t *testing.T, suffix string) string {
	t.Helper()
	codeRepo := database.NewRegistrationCodeRepositoryPGX(testPool)
	code := fmt.Sprintf("t%03d-%s", codeSeq.Add(1), suffix)
	rc, err := codeRepo.Create(context.Background(), "", code)
	require.NoError(t, err)
	return rc.Code
}

func registerBody(email, password, name, code string) []byte {
	body, _ := json.Marshal(domain.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
		Code:     code,
	})
	return body
}

func executeRegister(
	handler *api.MemberHandler,
	email, password, name, code string,
) *httptest.ResponseRecorder {
	return executeRequest(http.MethodPost, "/api/members/register", registerBody(email, password, name, code), handler.RegisterMember)
}

func TestHandler_Register(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()
	code := seedCode(t, "valid")

	w := executeRegister(handler, "user@example.com", "password123", "John Doe", code)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp domain.MemberResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "user@example.com", resp.Email)
	assert.Equal(t, "John Doe", resp.Name)
}

func TestHandler_Register_MissingCode(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	w := executeRegister(handler, "nocode@example.com", "password", "No Code", "")
	assert.Equal(t, http.StatusCreated, w.Code)
	var resp domain.MemberResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "customer", resp.MemberType)
}

func TestHandler_Register_InvalidCode(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	w := executeRegister(handler, "bad@example.com", "password", "Bad", "does-not-exist")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Register_UsedCode(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()
	code := seedCode(t, "single")

	w := executeRegister(handler, "first@example.com", "password", "First", code)
	require.Equal(t, http.StatusCreated, w.Code)

	w = executeRegister(handler, "second@example.com", "password", "Second", code)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Register_DuplicateEmail(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	w := executeRegister(handler, "dup@example.com", "password123", "User", seedCode(t, "a"))
	assert.Equal(t, http.StatusCreated, w.Code)

	w = executeRegister(handler, "dup@example.com", "password123", "User", seedCode(t, "b"))
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandler_Login(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	w := executeRegister(handler, "login@example.com", "mypassword", "Login User", seedCode(t, "login"))
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

	w := executeRegister(handler, "wrong@example.com", "correctpw", "User", seedCode(t, "wrong"))
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
	_, sessionCache, handler := setupMemberHandler()

	w := executeRegister(handler, "me@example.com", "password", "Me User", seedCode(t, "me"))
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody, _ := json.Marshal(domain.LoginRequest{
		Email:    "me@example.com",
		Password: "password",
	})
	w = executeRequest(http.MethodPost, "/api/members/login", loginBody, handler.LoginMember)
	require.Equal(t, http.StatusOK, w.Code)

	var loginResp domain.LoginResponse
	err := json.NewDecoder(w.Body).Decode(&loginResp)
	require.NoError(t, err)

	member, err := sessionCache.GetByKey(context.Background(), w.Result().Cookies()[0].Value)
	require.NoError(t, err)
	require.NotNil(t, member)

	req := httptest.NewRequest(http.MethodGet, "/api/members/me", nil)
	req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{
		ID:    loginResp.Member.ID,
		Email: loginResp.Member.Email,
		Name:  loginResp.Member.Name,
	}))
	w = httptest.NewRecorder()
	handler.GetCurrentMember(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.MemberResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
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
	_, sessionCache, handler := setupMemberHandler()

	w := executeRegister(handler, "expired@example.com", "password", "Expired User", seedCode(t, "expired"))
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody, _ := json.Marshal(domain.LoginRequest{
		Email:    "expired@example.com",
		Password: "password",
	})
	w = executeRequest(http.MethodPost, "/api/members/login", loginBody, handler.LoginMember)
	require.Equal(t, http.StatusOK, w.Code)

	sessionCookie := w.Result().Cookies()[0]

	sessionCache.Delete(context.Background(), sessionCookie.Value)

	req := httptest.NewRequest(http.MethodGet, "/api/members/me", nil)
	w = httptest.NewRecorder()
	handler.GetCurrentMember(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_UpdateMember(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	w := executeRegister(handler, "update@example.com", "password", "Original Name", seedCode(t, "update"))
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody, _ := json.Marshal(domain.LoginRequest{
		Email:    "update@example.com",
		Password: "password",
	})
	w = executeRequest(http.MethodPost, "/api/members/login", loginBody, handler.LoginMember)
	require.Equal(t, http.StatusOK, w.Code)

	var loginResp domain.LoginResponse
	err := json.NewDecoder(w.Body).Decode(&loginResp)
	require.NoError(t, err)

	memberCtx := api.ContextWithMember(context.Background(), &domain.Member{
		ID:    loginResp.Member.ID,
		Email: loginResp.Member.Email,
		Name:  loginResp.Member.Name,
	})

	t.Run("successful update", func(t *testing.T) {
		updateBody, _ := json.Marshal(domain.UpdateMemberRequest{
			Email: "updated@example.com",
			Name:  "Updated Name",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/members/update", bytes.NewReader(updateBody))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(memberCtx)
		w = httptest.NewRecorder()
		handler.UpdateMember(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.MemberResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, "updated@example.com", resp.Email)
		assert.Equal(t, "Updated Name", resp.Name)
	})

	t.Run("update without auth", func(t *testing.T) {
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
		req = req.WithContext(memberCtx)
		w = httptest.NewRecorder()
		handler.UpdateMember(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("update email to existing email", func(t *testing.T) {
		w := executeRegister(handler, "existing@example.com", "password", "Existing User", seedCode(t, "existing"))
		require.Equal(t, http.StatusCreated, w.Code)

		updateBody, _ := json.Marshal(domain.UpdateMemberRequest{
			Email: "existing@example.com",
			Name:  "Updated Name",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/members/update", bytes.NewReader(updateBody))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(memberCtx)
		w = httptest.NewRecorder()
		handler.UpdateMember(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestHandler_SessionIdleExpiry(t *testing.T) {
	defer cleanupMembers(t)
	memberRepo, sessionCache, handler := setupMemberHandler()

	w := executeRegister(handler, "idle@example.com", "password", "Idle User", seedCode(t, "idle"))
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody, _ := json.Marshal(domain.LoginRequest{
		Email:    "idle@example.com",
		Password: "password",
	})
	w = executeRequest(http.MethodPost, "/api/members/login", loginBody, handler.LoginMember)
	require.Equal(t, http.StatusOK, w.Code)

	sessionCookie := w.Result().Cookies()[0]
	require.NotNil(t, sessionCookie)

	session, err := sessionCache.GetByKey(context.Background(), sessionCookie.Value)
	require.NoError(t, err)
	require.NotNil(t, session)

	sessionCache.Delete(context.Background(), sessionCookie.Value)

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":"true"}`))
	})

	auth := api.AuthMiddleware(sessionCache, usecase.NewMemberService(memberRepo, sessionCache))
	req := httptest.NewRequest(http.MethodGet, "/api/members/me", nil)
	req.AddCookie(sessionCookie)
	w = httptest.NewRecorder()
	auth(dummyHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_Logout(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	w := executeRegister(handler, "logout@example.com", "password", "Logout User", seedCode(t, "logout"))
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

func TestHandler_Login_BadUser(t *testing.T) {
	defer cleanupMembers(t)
	_, _, handler := setupMemberHandler()

	loginBody, _ := json.Marshal(domain.LoginRequest{
		Email:    "nobody@example.com",
		Password: "password",
	})
	w := executeRequest(http.MethodPost, "/api/members/login", loginBody, handler.LoginMember)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
