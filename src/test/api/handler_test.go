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
	for _, table := range []string{"read_receipts", "chat_messages", "chat_room_members", "chat_rooms", "announcements", "inventory_items", "inventories", "members", "product_prices", "product_details", "products"} {
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS "+table+" CASCADE")
	}
	_, _ = pool.Exec(ctx, "DROP FUNCTION IF EXISTS create_member, get_member_by_email, get_member_by_id, update_member, create_product, list_products, get_product_by_id, update_product, delete_product, create_product_detail, get_product_detail_by_product, update_product_detail, create_product_price, get_product_price_by_id, list_product_prices_by_detail, update_product_price, create_inventory, get_inventory_by_id, get_inventory_by_price_id, list_inventories, update_inventory, delete_inventory, create_inventory_item, get_inventory_item_by_id, list_inventory_items, update_inventory_item, delete_inventory_item, create_announcement, get_announcement_by_id, list_announcements, count_announcements, update_announcement, delete_announcement, create_chat_room, add_chat_room_members, get_chat_room_by_id, list_chat_rooms_by_member, update_chat_room, delete_chat_room, remove_chat_room_member, send_message, list_messages, delete_message, mark_message_read, get_message_read_by, count_room_unread CASCADE")
}

func runMigration(ctx context.Context, pool *pgxpool.Pool) {
	for _, file := range []string{
		"../../../db/migrations/001_create_products.up.sql",
		"../../../db/migrations/002_create_members.up.sql",
		"../../../db/migrations/003_add_member_id_to_products.up.sql",
		"../../../db/migrations/004_create_inventory.up.sql",
		"../../../db/migrations/005_simplify_inventories.up.sql",
		"../../../db/migrations/006_create_functions.up.sql",
		"../../../db/migrations/007_add_inventory_id_to_price_functions.up.sql",
		"../../../db/migrations/008_create_announcements.up.sql",
		"../../../db/migrations/009_set_not_null.up.sql",
		"../../../db/migrations/010_create_chat.up.sql",
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
	_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE products CASCADE")
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
				Name:  "Test Product",
				Price: 19.99,
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
			Name:  "Product A",
			Price: 10.0,
		},
	)
	repo.Create(
		context.Background(),
		&domain.Product{
			Name:  "Product B",
			Price: 20.0,
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
		Name:  "Test",
		Price: 15.0,
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
				map[string]string{"productId": tt.id},
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
		Name:  "Original",
		Price: 10.0,
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
				Name:  "Updated",
				Price: 25.0,
			},
			wantStatus: http.StatusOK,
			wantName:   "Updated",
		},
		{
			name: "update non-existent product",
			id:   "non-existent-id",
			body: domain.UpdateProductRequest{
				Name:  "Nope",
				Price: 0,
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
				map[string]string{"productId": tt.id},
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

func TestHandler_CreateDetail(t *testing.T) {
	defer cleanupProducts(t)
	repo, memberRepo, sessionCache, handler := setupTestHandler()
	member := createAuthMember(t, memberRepo, sessionCache)

	product := domain.Product{
		Name: "Test Product",
	}
	err := repo.Create(context.Background(), &product)
	require.NoError(t, err)

	tests := []struct {
		name       string
		id         string
		body       any
		wantStatus int
	}{
		{
			name: "valid detail",
			id:   product.ID,
			body: domain.CreateDetailRequest{
				Introduction:      "產品介紹",
				UsageInstructions: "使用說明",
				ReturnPolicy:      "退貨說明",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "non-existent product",
			id:   "00000000-0000-0000-0000-000000000000",
			body: domain.CreateDetailRequest{
				Introduction: "測試",
			},
			wantStatus: http.StatusInternalServerError,
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

			req := httptest.NewRequest(http.MethodPost, "/api/products/"+tt.id+"/details", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"productId": tt.id})
			req = req.WithContext(api.ContextWithMember(req.Context(), member))
			w := httptest.NewRecorder()
			handler.CreateDetail(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusCreated {
				var resp domain.DetailResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, "產品介紹", resp.Detail.Introduction)
				assert.NotEmpty(t, resp.Detail.ID)
			}
		})
	}
}

func TestHandler_CreateDetail_Unauthorized(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, handler := setupTestHandler()

	body, _ := json.Marshal(domain.CreateDetailRequest{
		Introduction: "No Auth Detail",
	})

	w := executeRequestWithVars(
		http.MethodPost,
		"/api/products/some-id/details",
		body,
		map[string]string{"productId": "some-id"},
		handler.CreateDetail,
	)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_CreatePrice(t *testing.T) {
	defer cleanupProducts(t)
	repo, memberRepo, sessionCache, handler := setupTestHandler()
	member := createAuthMember(t, memberRepo, sessionCache)

	product := domain.Product{
		Name: "Test Product",
	}
	err := repo.Create(context.Background(), &product)
	require.NoError(t, err)

	detail := domain.ProductDetail{
		ProductID: product.ID,
	}
	err = repo.CreateDetail(context.Background(), &detail)
	require.NoError(t, err)

	tests := []struct {
		name       string
		productID  string
		detailID   string
		body       any
		wantStatus int
	}{
		{
			name:      "valid price",
			productID: product.ID,
			detailID:  detail.ID,
			body: domain.CreatePriceRequest{
				Label:     "成人票",
				Amount:    100.00,
				Currency:  "TWD",
				SortOrder: 1,
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:      "non-existent detail",
			productID: product.ID,
			detailID:  "00000000-0000-0000-0000-000000000000",
			body: domain.CreatePriceRequest{
				Label:  "Ghost",
				Amount: 0,
			},
			wantStatus: http.StatusInternalServerError,
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

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/products/"+tt.productID+"/details/"+tt.detailID+"/prices",
				bytes.NewReader(bodyBytes),
			)
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"productId": tt.productID, "detailId": tt.detailID})
			req = req.WithContext(api.ContextWithMember(req.Context(), member))
			w := httptest.NewRecorder()
			handler.CreatePrice(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusCreated {
				var resp domain.PriceResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.detailID, resp.Price.ProductDetailID)
				assert.NotEmpty(t, resp.Price.ID)
			}
		})
	}
}

func TestHandler_CreatePrice_Unauthorized(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, handler := setupTestHandler()

	body, _ := json.Marshal(domain.CreatePriceRequest{
		Label:  "No Auth Price",
		Amount: 50,
	})

	w := executeRequestWithVars(
		http.MethodPost,
		"/api/products/some-id/details/some-did/prices",
		body,
		map[string]string{"productId": "some-id", "detailId": "some-did"},
		handler.CreatePrice,
	)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_GetDetail(t *testing.T) {
	defer cleanupProducts(t)
	repo, _, _, handler := setupTestHandler()

	product := domain.Product{Name: "Test"}
	err := repo.Create(context.Background(), &product)
	require.NoError(t, err)

	detail := domain.ProductDetail{
		ProductID:         product.ID,
		Introduction:      "介紹",
		UsageInstructions: "說明",
		ReturnPolicy:      "退貨",
	}
	err = repo.CreateDetail(context.Background(), &detail)
	require.NoError(t, err)

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "existing detail",
			id:         product.ID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existent product",
			id:         "00000000-0000-0000-0000-000000000000",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := executeRequestWithVars(
				http.MethodGet,
				"/api/products/"+tt.id+"/detail",
				nil,
				map[string]string{"productId": tt.id},
				handler.GetDetail,
			)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp domain.DetailResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, detail.ID, resp.Detail.ID)
				assert.Equal(t, "介紹", resp.Detail.Introduction)
			}
		})
	}
}

func TestHandler_UpdateDetail(t *testing.T) {
	defer cleanupProducts(t)
	repo, memberRepo, sessionCache, handler := setupTestHandler()
	member := createAuthMember(t, memberRepo, sessionCache)

	product := domain.Product{Name: "Test"}
	err := repo.Create(context.Background(), &product)
	require.NoError(t, err)

	detail := domain.ProductDetail{
		ProductID: product.ID,
	}
	err = repo.CreateDetail(context.Background(), &detail)
	require.NoError(t, err)

	tests := []struct {
		name       string
		id         string
		body       any
		wantStatus int
	}{
		{
			name: "update existing detail",
			id:   product.ID,
			body: domain.UpdateDetailRequest{
				Introduction:      "新介紹",
				UsageInstructions: "新說明",
				ReturnPolicy:      "新退貨",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "non-existent product",
			id:   "00000000-0000-0000-0000-000000000000",
			body: domain.UpdateDetailRequest{
				Introduction: "無",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid json",
			id:         product.ID,
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

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/products/"+tt.id+"/detail/update",
				bytes.NewReader(bodyBytes),
			)
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"productId": tt.id})
			req = req.WithContext(api.ContextWithMember(req.Context(), member))
			w := httptest.NewRecorder()
			handler.UpdateDetail(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp domain.DetailResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, detail.ID, resp.Detail.ID)
				assert.Equal(t, "新介紹", resp.Detail.Introduction)
			}
		})
	}
}

func TestHandler_UpdateDetail_Unauthorized(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, handler := setupTestHandler()

	body, _ := json.Marshal(domain.UpdateDetailRequest{
		Introduction: "No Auth",
	})

	w := executeRequestWithVars(
		http.MethodPost,
		"/api/products/some-id/detail/update",
		body,
		map[string]string{"productId": "some-id"},
		handler.UpdateDetail,
	)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_ListPrices(t *testing.T) {
	defer cleanupProducts(t)
	repo, _, _, handler := setupTestHandler()

	product := domain.Product{Name: "Test"}
	err := repo.Create(context.Background(), &product)
	require.NoError(t, err)

	detail := domain.ProductDetail{ProductID: product.ID}
	err = repo.CreateDetail(context.Background(), &detail)
	require.NoError(t, err)

	repo.CreatePrice(context.Background(), &domain.ProductPrice{
		ProductDetailID: detail.ID, Label: "A", Amount: 10, SortOrder: 1,
	})
	repo.CreatePrice(context.Background(), &domain.ProductPrice{
		ProductDetailID: detail.ID, Label: "B", Amount: 20, SortOrder: 2,
	})

	t.Run("list prices", func(t *testing.T) {
		w := executeRequestWithVars(
			http.MethodGet,
			"/api/products/"+product.ID+"/detail/prices",
			nil,
			map[string]string{"productId": product.ID},
			handler.ListPrices,
		)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.PriceListResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Len(t, resp.Prices, 2)
	})

	t.Run("non-existent product", func(t *testing.T) {
		w := executeRequestWithVars(
			http.MethodGet,
			"/api/products/00000000-0000-0000-0000-000000000000/detail/prices",
			nil,
			map[string]string{"productId": "00000000-0000-0000-0000-000000000000"},
			handler.ListPrices,
		)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandler_GetPrice(t *testing.T) {
	defer cleanupProducts(t)
	repo, _, _, handler := setupTestHandler()

	product := domain.Product{Name: "Test"}
	err := repo.Create(context.Background(), &product)
	require.NoError(t, err)

	detail := domain.ProductDetail{ProductID: product.ID}
	err = repo.CreateDetail(context.Background(), &detail)
	require.NoError(t, err)

	price := domain.ProductPrice{
		ProductDetailID: detail.ID,
		Label:           "成人票",
		Amount:          100,
		SortOrder:       1,
	}
	err = repo.CreatePrice(context.Background(), &price)
	require.NoError(t, err)

	tests := []struct {
		name       string
		productID  string
		priceID    string
		wantStatus int
	}{
		{
			name:       "existing price",
			productID:  product.ID,
			priceID:    price.ID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existent price",
			productID:  product.ID,
			priceID:    "00000000-0000-0000-0000-000000000000",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := executeRequestWithVars(
				http.MethodGet,
				"/api/products/"+tt.productID+"/detail/prices/"+tt.priceID,
				nil,
				map[string]string{"productId": tt.productID, "priceId": tt.priceID},
				handler.GetPrice,
			)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp domain.PriceResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, price.ID, resp.Price.ID)
				assert.Equal(t, "成人票", resp.Price.Label)
			}
		})
	}
}

func TestHandler_UpdatePrice(t *testing.T) {
	defer cleanupProducts(t)
	repo, memberRepo, sessionCache, handler := setupTestHandler()
	member := createAuthMember(t, memberRepo, sessionCache)

	product := domain.Product{Name: "Test"}
	err := repo.Create(context.Background(), &product)
	require.NoError(t, err)

	detail := domain.ProductDetail{ProductID: product.ID}
	err = repo.CreateDetail(context.Background(), &detail)
	require.NoError(t, err)

	price := domain.ProductPrice{
		ProductDetailID: detail.ID,
		Label:           "原價",
		Amount:          100,
		SortOrder:       1,
	}
	err = repo.CreatePrice(context.Background(), &price)
	require.NoError(t, err)

	tests := []struct {
		name       string
		productID  string
		priceID    string
		body       any
		wantStatus int
	}{
		{
			name:      "update existing price",
			productID: product.ID,
			priceID:   price.ID,
			body: domain.UpdatePriceRequest{
				Label:     "特價",
				Amount:    80,
				Currency:  "USD",
				SortOrder: 2,
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "non-existent price",
			productID: product.ID,
			priceID:   "00000000-0000-0000-0000-000000000000",
			body: domain.UpdatePriceRequest{
				Label:  "Ghost",
				Amount: 0,
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid json",
			productID:  product.ID,
			priceID:    price.ID,
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

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/products/"+tt.productID+"/detail/prices/"+tt.priceID+"/update",
				bytes.NewReader(bodyBytes),
			)
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"productId": tt.productID, "priceId": tt.priceID})
			req = req.WithContext(api.ContextWithMember(req.Context(), member))
			w := httptest.NewRecorder()
			handler.UpdatePrice(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp domain.PriceResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, price.ID, resp.Price.ID)
				assert.Equal(t, "特價", resp.Price.Label)
			}
		})
	}
}

func TestHandler_UpdatePrice_Unauthorized(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, handler := setupTestHandler()

	body, _ := json.Marshal(domain.UpdatePriceRequest{
		Label:  "No Auth",
		Amount: 50,
	})

	w := executeRequestWithVars(
		http.MethodPost,
		"/api/products/some-id/detail/prices/some-pid/update",
		body,
		map[string]string{"productId": "some-id", "priceId": "some-pid"},
		handler.UpdatePrice,
	)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_DeleteProduct(t *testing.T) {
	defer cleanupProducts(t)
	repo, _, _, handler := setupTestHandler()

	created := domain.Product{
		Name:  "To Delete",
		Price: 5.0,
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
				map[string]string{"productId": tt.id},
				handler.DeleteProduct,
			)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
