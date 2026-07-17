//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"backend/src/api"
	"backend/src/database"
	"backend/src/domain"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://root:root123@localhost:5432/productdb_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}

	dropTables(ctx, pool)
	runMigration(ctx, pool)

	testPool = pool

	code := m.Run()

	pool.Close()
	os.Exit(code)
}

func dropTables(ctx context.Context, pool *pgxpool.Pool) {
	for _, table := range []string{"members", "product_prices", "product_details", "products"} {
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS "+table+" CASCADE")
	}
}

func runMigration(ctx context.Context, pool *pgxpool.Pool) {
	for _, file := range []string{
		"../../../db/migrations/001_create_products.sql",
		"../../../db/migrations/002_create_members.sql",
		"../../../db/migrations/003_add_member_id_to_products.sql",
	} {
		schema, err := os.ReadFile(file)
		if err != nil {
			panic("failed to read migration file: " + err.Error())
		}
		_, err = pool.Exec(ctx, string(schema))
		if err != nil {
			panic("failed to run migration: " + err.Error())
		}
	}
}

func setupTestHandler() (*database.ProductRepositoryPGX, *database.MemberRepositoryPGX, *database.SessionCache, *api.ProductHandler) {
	repo := database.NewProductRepositoryPGX(testPool)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	sessionCache := database.NewSessionCache(24 * time.Hour)
	handler := api.NewProductHandler(repo)
	return repo, memberRepo, sessionCache, handler
}

func cleanupProducts(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE products, members CASCADE")
	require.NoError(t, err)
}

func createAuthMember(t *testing.T, memberRepo *database.MemberRepositoryPGX, sessionCache *database.SessionCache) *domain.Member {
	t.Helper()

	member := domain.Member{
		Email:    "test-" + t.Name() + "@example.com",
		Password: "password",
		Name:     "Test User",
	}
	err := memberRepo.Create(context.Background(), &member)
	require.NoError(t, err)

	session := domain.Session{MemberID: member.ID}
	err = sessionCache.Create(context.Background(), &session)
	require.NoError(t, err)

	return &member
}

func executeRequest(
	method, path string,
	body []byte,
	handler http.HandlerFunc,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func executeRequestWithVars(
	method, path string,
	body []byte,
	vars map[string]string,
	handler http.HandlerFunc,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestHandler_CreateProduct(t *testing.T) {
	defer cleanupProducts(t)
	repo, memberRepo, sessionCache, handler := setupTestHandler()
	member := createAuthMember(t, memberRepo, sessionCache)

	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantName   string
	}{
		{
			name: "valid product",
			body: domain.CreateProductRequest{
				Name:        "Test Product",
				Description: "A test product",
				Price:       19.99,
			},
			wantStatus: http.StatusCreated,
			wantName:   "Test Product",
		},
		{
			name:       "invalid json",
			body:       "{invalid}",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			default:
				bodyBytes, _ = json.Marshal(b)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(api.ContextWithMember(req.Context(), member))
			w := httptest.NewRecorder()
			handler.CreateProduct(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusCreated {
				var resp domain.ProductResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantName, resp.Product.Name)
				assert.NotEmpty(t, resp.Product.ID)
				assert.NotEmpty(t, resp.Product.CreatedBy)
			}
		})
	}

	_ = repo
}

func TestHandler_CreateProduct_Unauthorized(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, handler := setupTestHandler()

	body, _ := json.Marshal(domain.CreateProductRequest{
		Name: "No Auth Product",
	})

	w := executeRequest(http.MethodPost, "/api/products", body, handler.CreateProduct)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_ListProducts(t *testing.T) {
	defer cleanupProducts(t)
	repo, _, _, handler := setupTestHandler()

	repo.Create(
		context.Background(),
		&domain.Product{
			Name:        "Product A",
			Description: "Desc A",
			Price:       10.0,
		},
	)
	repo.Create(
		context.Background(),
		&domain.Product{
			Name:        "Product B",
			Description: "Desc B",
			Price:       20.0,
		},
	)

	w := executeRequest(
		http.MethodGet,
		"/api/products",
		nil,
		handler.ListProducts,
	)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.ProductListResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Products, 2)
}

func TestHandler_ListProducts_Empty(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, handler := setupTestHandler()

	w := executeRequest(
		http.MethodGet,
		"/api/products",
		nil,
		handler.ListProducts,
	)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.ProductListResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Empty(t, resp.Products)
}

func TestHandler_GetProduct(t *testing.T) {
	defer cleanupProducts(t)
	repo, _, _, handler := setupTestHandler()

	created := domain.Product{
		Name:        "Test",
		Description: "Desc",
		Price:       15.0,
	}
	repo.Create(context.Background(), &created)

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "existing product",
			id:         created.ID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existent product",
			id:         "non-existent-id",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := executeRequestWithVars(
				http.MethodGet,
				"/api/products/"+tt.id,
				nil,
				map[string]string{"id": tt.id},
				handler.GetProduct,
			)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp domain.ProductResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, created.ID, resp.Product.ID)
				assert.Equal(t, "Test", resp.Product.Name)
			}
		})
	}
}

func TestHandler_UpdateProduct(t *testing.T) {
	defer cleanupProducts(t)
	repo, _, _, handler := setupTestHandler()

	created := domain.Product{
		Name:        "Original",
		Description: "Original desc",
		Price:       10.0,
	}
	repo.Create(context.Background(), &created)

	tests := []struct {
		name       string
		id         string
		body       any
		wantStatus int
		wantName   string
	}{
		{
			name: "update existing product",
			id:   created.ID,
			body: domain.UpdateProductRequest{
				Name:        "Updated",
				Description: "Updated desc",
				Price:       25.0,
			},
			wantStatus: http.StatusOK,
			wantName:   "Updated",
		},
		{
			name: "update non-existent product",
			id:   "non-existent-id",
			body: domain.UpdateProductRequest{
				Name:        "Nope",
				Description: "Nope",
				Price:       0,
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid json",
			id:         created.ID,
			body:       "{invalid}",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			default:
				bodyBytes, _ = json.Marshal(b)
			}

			w := executeRequestWithVars(
				http.MethodPost,
				"/api/products/"+tt.id+"/update",
				bodyBytes,
				map[string]string{"id": tt.id},
				handler.UpdateProduct,
			)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp domain.ProductResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantName, resp.Product.Name)
			}
		})
	}
}

func TestHandler_DeleteProduct(t *testing.T) {
	defer cleanupProducts(t)
	repo, _, _, handler := setupTestHandler()

	created := domain.Product{
		Name:        "To Delete",
		Description: "Desc",
		Price:       5.0,
	}
	repo.Create(context.Background(), &created)

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "delete existing product",
			id:         created.ID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "delete non-existent product",
			id:         created.ID,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := executeRequestWithVars(
				http.MethodPost,
				"/api/products/"+tt.id+"/delete",
				nil,
				map[string]string{"id": tt.id},
				handler.DeleteProduct,
			)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
