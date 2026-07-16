package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/src/api"
	"backend/src/database"
	"backend/src/domain"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupTestHandler() (*database.InMemoryRepository, *api.ProductHandler) {
	repo := database.NewInMemoryRepository()
	handler := api.NewProductHandler(repo)
	return repo, handler
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
	_, handler := setupTestHandler()

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

			w := executeRequest(
				http.MethodPost,
				"/api/products",
				bodyBytes,
				handler.CreateProduct,
			)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusCreated {
				var resp domain.ProductResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantName, resp.Product.Name)
				assert.NotEmpty(t, resp.Product.ID)
			}
		})
	}
}

func TestHandler_ListProducts(t *testing.T) {
	repo, handler := setupTestHandler()

	repo.Create(
		nil,
		&domain.Product{
			Name:        "Product A",
			Description: "Desc A",
			Price:       10.0,
		},
	)
	repo.Create(
		nil,
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
	_, handler := setupTestHandler()

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
	repo, handler := setupTestHandler()

	created := domain.Product{
		Name:        "Test",
		Description: "Desc",
		Price:       15.0,
	}
	repo.Create(nil, &created)

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
	repo, handler := setupTestHandler()

	created := domain.Product{
		Name:        "Original",
		Description: "Original desc",
		Price:       10.0,
	}
	repo.Create(nil, &created)

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
	repo, handler := setupTestHandler()

	created := domain.Product{
		Name:        "To Delete",
		Description: "Desc",
		Price:       5.0,
	}
	repo.Create(nil, &created)

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
