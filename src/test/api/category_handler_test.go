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

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var categorySeq atomic.Uint64

func setupCategoryRouter() (*database.MemberRepositoryPGX, *session.SessionCache, *database.CategoryRepositoryPGX, *mux.Router) {
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	sessionCache := session.NewSessionCache(time.Hour)
	repo := database.NewCategoryRepositoryPGX(testPool)
	handler := composeCategoryHandler(repo)
	auth := api.AuthMiddleware(sessionCache, usecase.NewMemberService(memberRepo, sessionCache))

	r := mux.NewRouter()
	r.HandleFunc("/api/categories", handler.ListCategories).Methods("GET")
	r.Handle("/api/categories", auth(http.HandlerFunc(handler.CreateCategory))).Methods("POST")
	r.Handle("/api/categories/{id}/update", auth(http.HandlerFunc(handler.UpdateCategory))).Methods("POST")
	r.Handle("/api/categories/{id}/delete", auth(http.HandlerFunc(handler.DeleteCategory))).Methods("POST")

	return memberRepo, sessionCache, repo, r
}

func cleanupCategories(t *testing.T) {
	t.Helper()
	require.NoError(t, testHarness.Reset(context.Background()))
}

func doCategoryReq(t *testing.T, r *mux.Router, method, path string, body []byte, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func freshCategoryName() string {
	return fmt.Sprintf("cat-%d", categorySeq.Add(1))
}

func TestCategory_Create(t *testing.T) {
	defer cleanupCategories(t)
	memberRepo, sessionCache, _, r := setupCategoryRouter()
	cookie := authCookie(t, memberRepo, sessionCache, "member")

	body, _ := json.Marshal(domain.CreateCategoryRequest{Name: "electronics"})
	w := doCategoryReq(t, r, "POST", "/api/categories", body, cookie)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp domain.CategoryResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "electronics", resp.Category.Name)
	assert.NotEmpty(t, resp.Category.ID)
}

func TestCategory_CreateEmptyName(t *testing.T) {
	defer cleanupCategories(t)
	memberRepo, sessionCache, _, r := setupCategoryRouter()
	cookie := authCookie(t, memberRepo, sessionCache, "member")

	body, _ := json.Marshal(domain.CreateCategoryRequest{Name: "   "})
	w := doCategoryReq(t, r, "POST", "/api/categories", body, cookie)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategory_CreateDuplicate(t *testing.T) {
	defer cleanupCategories(t)
	memberRepo, sessionCache, _, r := setupCategoryRouter()
	cookie := authCookie(t, memberRepo, sessionCache, "member")

	name := freshCategoryName()
	_, err := database.NewCategoryRepositoryPGX(testPool).Create(context.Background(), name)
	require.NoError(t, err)

	body, _ := json.Marshal(domain.CreateCategoryRequest{Name: name})
	w := doCategoryReq(t, r, "POST", "/api/categories", body, cookie)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCategory_List(t *testing.T) {
	defer cleanupCategories(t)
	memberRepo, sessionCache, _, r := setupCategoryRouter()
	cookie := authCookie(t, memberRepo, sessionCache, "member")

	repo := database.NewCategoryRepositoryPGX(testPool)
	first, err := repo.Create(context.Background(), freshCategoryName())
	require.NoError(t, err)
	second, err := repo.Create(context.Background(), freshCategoryName())
	require.NoError(t, err)

	w := doCategoryReq(t, r, "GET", "/api/categories", nil, cookie)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.CategoryListResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Categories, 2)
	assert.Equal(t, second.ID, resp.Categories[0].ID)
	assert.Equal(t, first.ID, resp.Categories[1].ID)
}

func TestCategory_ListUnauthenticated(t *testing.T) {
	defer cleanupCategories(t)
	_, _, _, r := setupCategoryRouter()

	w := doCategoryReq(t, r, "GET", "/api/categories", nil, nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCategory_Update(t *testing.T) {
	defer cleanupCategories(t)
	memberRepo, sessionCache, _, r := setupCategoryRouter()
	cookie := authCookie(t, memberRepo, sessionCache, "member")

	repo := database.NewCategoryRepositoryPGX(testPool)
	created, err := repo.Create(context.Background(), "before")
	require.NoError(t, err)

	body, _ := json.Marshal(domain.UpdateCategoryRequest{Name: "after"})
	w := doCategoryReq(t, r, "POST", "/api/categories/"+created.ID+"/update", body, cookie)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.CategoryResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, created.ID, resp.Category.ID)
	assert.Equal(t, "after", resp.Category.Name)
}

func TestCategory_UpdateNotFound(t *testing.T) {
	defer cleanupCategories(t)
	memberRepo, sessionCache, _, r := setupCategoryRouter()
	cookie := authCookie(t, memberRepo, sessionCache, "member")

	body, _ := json.Marshal(domain.UpdateCategoryRequest{Name: "ghost"})
	w := doCategoryReq(t, r, "POST", "/api/categories/00000000-0000-0000-0000-000000000000/update", body, cookie)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCategory_UpdateDuplicate(t *testing.T) {
	defer cleanupCategories(t)
	memberRepo, sessionCache, _, r := setupCategoryRouter()
	cookie := authCookie(t, memberRepo, sessionCache, "member")

	repo := database.NewCategoryRepositoryPGX(testPool)
	name := freshCategoryName()
	_, err := repo.Create(context.Background(), name)
	require.NoError(t, err)
	dupTarget, err := repo.Create(context.Background(), freshCategoryName())
	require.NoError(t, err)

	body, _ := json.Marshal(domain.UpdateCategoryRequest{Name: name})
	w := doCategoryReq(t, r, "POST", "/api/categories/"+dupTarget.ID+"/update", body, cookie)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCategory_UpdateEmptyName(t *testing.T) {
	defer cleanupCategories(t)
	memberRepo, sessionCache, repo, r := setupCategoryRouter()
	cookie := authCookie(t, memberRepo, sessionCache, "member")

	created, err := repo.Create(context.Background(), "to-rename")
	require.NoError(t, err)

	body, _ := json.Marshal(domain.UpdateCategoryRequest{Name: " "})
	w := doCategoryReq(t, r, "POST", "/api/categories/"+created.ID+"/update", body, cookie)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategory_Delete(t *testing.T) {
	defer cleanupCategories(t)
	memberRepo, sessionCache, _, r := setupCategoryRouter()
	cookie := authCookie(t, memberRepo, sessionCache, "member")

	repo := database.NewCategoryRepositoryPGX(testPool)
	created, err := repo.Create(context.Background(), freshCategoryName())
	require.NoError(t, err)

	w := doCategoryReq(t, r, "POST", "/api/categories/"+created.ID+"/delete", nil, cookie)
	assert.Equal(t, http.StatusOK, w.Code)

	var count int
	_ = testPool.QueryRow(context.Background(), "SELECT count(*) FROM categories WHERE id = $1", created.ID).Scan(&count)
	assert.Equal(t, 0, count)
}

func TestCategory_DeleteNotFound(t *testing.T) {
	defer cleanupCategories(t)
	memberRepo, sessionCache, _, r := setupCategoryRouter()
	cookie := authCookie(t, memberRepo, sessionCache, "member")

	w := doCategoryReq(t, r, "POST", "/api/categories/00000000-0000-0000-0000-000000000000/delete", nil, cookie)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCategory_DeleteInUse(t *testing.T) {
	defer cleanupCategories(t)
	memberRepo, sessionCache, _, r := setupCategoryRouter()
	cookie := authCookie(t, memberRepo, sessionCache, "member")

	repo := database.NewCategoryRepositoryPGX(testPool)
	created, err := repo.Create(context.Background(), freshCategoryName())
	require.NoError(t, err)

	productRepo := database.NewProductRepositoryPGX(testPool)
	product := domain.Product{
		Name:       "Referencing Product",
		Price:      15.0,
		CategoryID: created.ID,
	}
	require.NoError(t, productRepo.Create(context.Background(), &product))

	w := doCategoryReq(t, r, "POST", "/api/categories/"+created.ID+"/delete", nil, cookie)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCategory_Unauthorized(t *testing.T) {
	defer cleanupCategories(t)
	_, _, _, r := setupCategoryRouter()

	body, _ := json.Marshal(domain.CreateCategoryRequest{Name: "nope"})
	w := doCategoryReq(t, r, "POST", "/api/categories", body, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
