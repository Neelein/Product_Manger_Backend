//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/src/api"
	"backend/src/database"
	"backend/src/domain"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupInventoryTest(t *testing.T) (*database.InventoryRepositoryPGX, *database.ProductRepositoryPGX, *database.MemberRepositoryPGX, *database.SessionCache, *api.InventoryHandler) {
	t.Helper()
	invRepo := database.NewInventoryRepositoryPGX(testPool)
	productRepo := database.NewProductRepositoryPGX(testPool)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	sessionCache := database.NewSessionCache(24 * time.Hour)
	handler := api.NewInventoryHandler(invRepo)
	return invRepo, productRepo, memberRepo, sessionCache, handler
}

func createTestPriceForHandler(t *testing.T, productRepo *database.ProductRepositoryPGX) domain.ProductPrice {
	t.Helper()
	product := domain.Product{Name: "Test Product", Price: 100}
	err := productRepo.Create(context.Background(), &product)
	require.NoError(t, err)

	detail := domain.ProductDetail{ProductID: product.ID}
	err = productRepo.CreateDetail(context.Background(), &detail)
	require.NoError(t, err)

	price := domain.ProductPrice{
		ProductDetailID: detail.ID,
		Label:           "Test Price",
		Amount:          100,
		SortOrder:       1,
	}
	err = productRepo.CreatePrice(context.Background(), &price)
	require.NoError(t, err)

	return price
}

func executeRequestWithVarsAndMember(
	method, path string,
	body []byte,
	vars map[string]string,
	member *domain.Member,
	handler http.HandlerFunc,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, vars)
	if member != nil {
		req = req.WithContext(api.ContextWithMember(req.Context(), member))
	}
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestInventoryHandler_CreateInventory(t *testing.T) {
	defer cleanupProducts(t)
	_, productRepo, memberRepo, sessionCache, handler := setupInventoryTest(t)
	member := createAuthMember(t, memberRepo, sessionCache)
	price := createTestPriceForHandler(t, productRepo)

	tests := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{
			name: "valid inventory",
			body: domain.CreateInventoryRequest{
				ProductPriceID: price.ID,
				Status:         "銷售中",
			},
			wantStatus: http.StatusCreated,
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

			req := httptest.NewRequest(http.MethodPost, "/api/inventories", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(api.ContextWithMember(req.Context(), member))
			w := httptest.NewRecorder()
			handler.CreateInventory(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusCreated {
				var resp domain.InventoryResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Inventory.ID)
			}
		})
	}
}

func TestInventoryHandler_CreateInventory_Unauthorized(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, _, handler := setupInventoryTest(t)

	body, _ := json.Marshal(domain.CreateInventoryRequest{
		ProductPriceID: "some-id",
	})

	w := executeRequest(http.MethodPost, "/api/inventories", body, handler.CreateInventory)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestInventoryHandler_ListInventories(t *testing.T) {
	defer cleanupProducts(t)
	invRepo, productRepo, _, _, handler := setupInventoryTest(t)
	price1 := createTestPriceForHandler(t, productRepo)
	price2 := createTestPriceForHandler(t, productRepo)

	invRepo.CreateInventory(context.Background(), &domain.Inventory{
		ProductPriceID: price1.ID,
	})
	invRepo.CreateInventory(context.Background(), &domain.Inventory{
		ProductPriceID: price2.ID,
	})

	w := executeRequest(http.MethodGet, "/api/inventories", nil, handler.ListInventories)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.InventoryListResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Inventories, 2)
}

func TestInventoryHandler_ListInventories_Empty(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, _, handler := setupInventoryTest(t)

	w := executeRequest(http.MethodGet, "/api/inventories", nil, handler.ListInventories)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.InventoryListResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Empty(t, resp.Inventories)
}

func TestInventoryHandler_GetInventory(t *testing.T) {
	defer cleanupProducts(t)
	invRepo, productRepo, _, _, handler := setupInventoryTest(t)
	price := createTestPriceForHandler(t, productRepo)

	created := domain.Inventory{
		ProductPriceID: price.ID,
	}
	invRepo.CreateInventory(context.Background(), &created)

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "existing inventory",
			id:         created.ID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existent inventory",
			id:         "00000000-0000-0000-0000-000000000000",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := executeRequestWithVars(
				http.MethodGet,
				"/api/inventories/"+tt.id,
				nil,
				map[string]string{"inventoryId": tt.id},
				handler.GetInventory,
			)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp domain.InventoryResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, created.ID, resp.Inventory.ID)
			}
		})
	}
}

func TestInventoryHandler_UpdateInventory(t *testing.T) {
	defer cleanupProducts(t)
	invRepo, productRepo, memberRepo, sessionCache, handler := setupInventoryTest(t)
	member := createAuthMember(t, memberRepo, sessionCache)
	price := createTestPriceForHandler(t, productRepo)

	created := domain.Inventory{
		ProductPriceID: price.ID,
	}
	invRepo.CreateInventory(context.Background(), &created)

	tests := []struct {
		name       string
		id         string
		body       any
		wantStatus int
	}{
		{
			name: "update status only",
			id:   created.ID,
			body: domain.UpdateInventoryRequest{
				Status: "完售",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "update non-existent inventory",
			id:   "00000000-0000-0000-0000-000000000000",
			body: domain.UpdateInventoryRequest{
				Status: "完售",
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

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/inventories/"+tt.id+"/update",
				bytes.NewReader(bodyBytes),
			)
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"inventoryId": tt.id})
			req = req.WithContext(api.ContextWithMember(req.Context(), member))
			w := httptest.NewRecorder()
			handler.UpdateInventory(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp domain.InventoryResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, "完售", resp.Inventory.Status)
			}
		})
	}
}

func TestInventoryHandler_UpdateInventory_Unauthorized(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, _, handler := setupInventoryTest(t)

	body, _ := json.Marshal(domain.UpdateInventoryRequest{
		Status: "完售",
	})

	w := executeRequestWithVars(
		http.MethodPost,
		"/api/inventories/some-id/update",
		body,
		map[string]string{"inventoryId": "some-id"},
		handler.UpdateInventory,
	)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestInventoryHandler_DeleteInventory(t *testing.T) {
	defer cleanupProducts(t)
	invRepo, productRepo, memberRepo, sessionCache, handler := setupInventoryTest(t)
	member := createAuthMember(t, memberRepo, sessionCache)
	price := createTestPriceForHandler(t, productRepo)

	created := domain.Inventory{
		ProductPriceID: price.ID,
	}
	invRepo.CreateInventory(context.Background(), &created)

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "delete existing",
			id:         created.ID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "delete non-existent",
			id:         created.ID,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := executeRequestWithVarsAndMember(
				http.MethodPost,
				"/api/inventories/"+tt.id+"/delete",
				nil,
				map[string]string{"inventoryId": tt.id},
				member,
				handler.DeleteInventory,
			)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestInventoryHandler_DeleteInventory_Unauthorized(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, _, handler := setupInventoryTest(t)

	w := executeRequestWithVars(
		http.MethodPost,
		"/api/inventories/some-id/delete",
		nil,
		map[string]string{"inventoryId": "some-id"},
		handler.DeleteInventory,
	)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestInventoryHandler_CreateItem(t *testing.T) {
	defer cleanupProducts(t)
	invRepo, productRepo, memberRepo, sessionCache, handler := setupInventoryTest(t)
	member := createAuthMember(t, memberRepo, sessionCache)
	price := createTestPriceForHandler(t, productRepo)

	inventory := domain.Inventory{
		ProductPriceID: price.ID,
	}
	invRepo.CreateInventory(context.Background(), &inventory)

	tests := []struct {
		name       string
		id         string
		body       any
		wantStatus int
	}{
		{
			name: "valid item",
			id:   inventory.ID,
			body: domain.CreateInventoryItemRequest{
				ItemCode:  "ITEM-001",
				Status:    "可用",
				Cost:      50.00,
				DateAdded: "2026-07-18",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid json",
			id:         inventory.ID,
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
				"/api/inventories/"+tt.id+"/items",
				bytes.NewReader(bodyBytes),
			)
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"inventoryId": tt.id})
			req = req.WithContext(api.ContextWithMember(req.Context(), member))
			w := httptest.NewRecorder()
			handler.CreateItem(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusCreated {
				var resp domain.InventoryItemResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, inventory.ID, resp.Item.InventoryID)
				assert.NotEmpty(t, resp.Item.ID)
			}
		})
	}
}

func TestInventoryHandler_CreateItem_Unauthorized(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, _, handler := setupInventoryTest(t)

	body, _ := json.Marshal(domain.CreateInventoryItemRequest{
		ItemCode: "No Auth",
	})

	w := executeRequestWithVars(
		http.MethodPost,
		"/api/inventories/some-id/items",
		body,
		map[string]string{"inventoryId": "some-id"},
		handler.CreateItem,
	)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestInventoryHandler_ListItems(t *testing.T) {
	defer cleanupProducts(t)
	invRepo, productRepo, _, _, handler := setupInventoryTest(t)
	price := createTestPriceForHandler(t, productRepo)

	inventory := domain.Inventory{
		ProductPriceID: price.ID,
	}
	invRepo.CreateInventory(context.Background(), &inventory)

	invRepo.CreateItem(context.Background(), &domain.InventoryItem{
		InventoryID: inventory.ID, ItemCode: "A",
	})
	invRepo.CreateItem(context.Background(), &domain.InventoryItem{
		InventoryID: inventory.ID, ItemCode: "B",
	})

	w := executeRequestWithVars(
		http.MethodGet,
		"/api/inventories/"+inventory.ID+"/items",
		nil,
		map[string]string{"inventoryId": inventory.ID},
		handler.ListItems,
	)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.InventoryItemListResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Items, 2)
}

func TestInventoryHandler_ListItems_Empty(t *testing.T) {
	defer cleanupProducts(t)
	invRepo, productRepo, _, _, handler := setupInventoryTest(t)
	price := createTestPriceForHandler(t, productRepo)

	inventory := domain.Inventory{
		ProductPriceID: price.ID,
	}
	invRepo.CreateInventory(context.Background(), &inventory)

	w := executeRequestWithVars(
		http.MethodGet,
		"/api/inventories/"+inventory.ID+"/items",
		nil,
		map[string]string{"inventoryId": inventory.ID},
		handler.ListItems,
	)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.InventoryItemListResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Empty(t, resp.Items)
}

func TestInventoryHandler_GetItem(t *testing.T) {
	defer cleanupProducts(t)
	invRepo, productRepo, _, _, handler := setupInventoryTest(t)
	price := createTestPriceForHandler(t, productRepo)

	inventory := domain.Inventory{
		ProductPriceID: price.ID,
	}
	invRepo.CreateInventory(context.Background(), &inventory)

	created := domain.InventoryItem{
		InventoryID: inventory.ID,
		ItemCode:    "ITEM-001",
		Status:      "可用",
		Cost:        50.00,
		DateAdded:   "2026-07-18",
	}
	invRepo.CreateItem(context.Background(), &created)

	tests := []struct {
		name         string
		inventoryID  string
		itemID       string
		wantStatus   int
	}{
		{
			name:         "existing item",
			inventoryID:  inventory.ID,
			itemID:       created.ID,
			wantStatus:   http.StatusOK,
		},
		{
			name:         "non-existent item",
			inventoryID:  inventory.ID,
			itemID:       "00000000-0000-0000-0000-000000000000",
			wantStatus:   http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := executeRequestWithVars(
				http.MethodGet,
				"/api/inventories/"+tt.inventoryID+"/items/"+tt.itemID,
				nil,
				map[string]string{"inventoryId": tt.inventoryID, "itemId": tt.itemID},
				handler.GetItem,
			)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp domain.InventoryItemResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, created.ID, resp.Item.ID)
				assert.Equal(t, "ITEM-001", resp.Item.ItemCode)
			}
		})
	}
}

func TestInventoryHandler_UpdateItem(t *testing.T) {
	defer cleanupProducts(t)
	invRepo, productRepo, memberRepo, sessionCache, handler := setupInventoryTest(t)
	member := createAuthMember(t, memberRepo, sessionCache)
	price := createTestPriceForHandler(t, productRepo)

	inventory := domain.Inventory{
		ProductPriceID: price.ID,
	}
	invRepo.CreateInventory(context.Background(), &inventory)

	created := domain.InventoryItem{
		InventoryID: inventory.ID,
		ItemCode:    "OLD-CODE",
		Status:      "可用",
	}
	invRepo.CreateItem(context.Background(), &created)

	tests := []struct {
		name         string
		inventoryID  string
		itemID       string
		body         any
		wantStatus   int
	}{
		{
			name:        "update existing item",
			inventoryID: inventory.ID,
			itemID:      created.ID,
			body: domain.UpdateInventoryItemRequest{
				ItemCode:  "NEW-CODE",
				Status:    "出售",
				Cost:      100.00,
				DateAdded: "2026-07-18",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "non-existent item",
			inventoryID: inventory.ID,
			itemID:      "00000000-0000-0000-0000-000000000000",
			body: domain.UpdateInventoryItemRequest{
				ItemCode: "Ghost",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:        "invalid json",
			inventoryID: inventory.ID,
			itemID:      created.ID,
			body:        "{invalid}",
			wantStatus:  http.StatusBadRequest,
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
				"/api/inventories/"+tt.inventoryID+"/items/"+tt.itemID+"/update",
				bytes.NewReader(bodyBytes),
			)
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"inventoryId": tt.inventoryID, "itemId": tt.itemID})
			req = req.WithContext(api.ContextWithMember(req.Context(), member))
			w := httptest.NewRecorder()
			handler.UpdateItem(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp domain.InventoryItemResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, "NEW-CODE", resp.Item.ItemCode)
				assert.Equal(t, "出售", resp.Item.Status)
			}
		})
	}
}

func TestInventoryHandler_UpdateItem_Unauthorized(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, _, handler := setupInventoryTest(t)

	body, _ := json.Marshal(domain.UpdateInventoryItemRequest{
		ItemCode: "No Auth",
	})

	w := executeRequestWithVars(
		http.MethodPost,
		"/api/inventories/some-id/items/some-iid/update",
		body,
		map[string]string{"inventoryId": "some-id", "itemId": "some-iid"},
		handler.UpdateItem,
	)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestInventoryHandler_DeleteItem(t *testing.T) {
	defer cleanupProducts(t)
	invRepo, productRepo, memberRepo, sessionCache, handler := setupInventoryTest(t)
	member := createAuthMember(t, memberRepo, sessionCache)
	price := createTestPriceForHandler(t, productRepo)

	inventory := domain.Inventory{
		ProductPriceID: price.ID,
	}
	invRepo.CreateInventory(context.Background(), &inventory)

	created := domain.InventoryItem{
		InventoryID: inventory.ID,
		ItemCode:    "TO-DELETE",
	}
	invRepo.CreateItem(context.Background(), &created)

	tests := []struct {
		name         string
		inventoryID  string
		itemID       string
		wantStatus   int
	}{
		{
			name:        "delete existing",
			inventoryID: inventory.ID,
			itemID:      created.ID,
			wantStatus:  http.StatusOK,
		},
		{
			name:        "delete non-existent",
			inventoryID: inventory.ID,
			itemID:      created.ID,
			wantStatus:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := executeRequestWithVarsAndMember(
				http.MethodPost,
				"/api/inventories/"+tt.inventoryID+"/items/"+tt.itemID+"/delete",
				nil,
				map[string]string{"inventoryId": tt.inventoryID, "itemId": tt.itemID},
				member,
				handler.DeleteItem,
			)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestInventoryHandler_DeleteItem_Unauthorized(t *testing.T) {
	defer cleanupProducts(t)
	_, _, _, _, handler := setupInventoryTest(t)

	w := executeRequestWithVars(
		http.MethodPost,
		"/api/inventories/some-id/items/some-iid/delete",
		nil,
		map[string]string{"inventoryId": "some-id", "itemId": "some-iid"},
		handler.DeleteItem,
	)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
