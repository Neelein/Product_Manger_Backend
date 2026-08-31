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

	api "backend/src/adapter/http"
	domain "backend/src/adapter/http"
	database "backend/src/adapter/postgres"
	"backend/src/adapter/session"
	"backend/src/usecase"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createVariantFixture(t *testing.T, repo *database.ProductRepositoryPGX) (domain.Product, domain.ProductDetail, domain.ProductPrice, domain.ProductOption, domain.ProductVariant) {
	t.Helper()
	product := domain.Product{Name: "Variant Product"}
	require.NoError(t, repo.Create(context.Background(), &product))
	detail := domain.ProductDetail{ProductID: product.ID}
	require.NoError(t, repo.CreateDetail(context.Background(), &detail))
	price := domain.ProductPrice{ProductDetailID: detail.ID, Label: "standard", Amount: 10}
	require.NoError(t, repo.CreatePrice(context.Background(), &price))
	option := domain.ProductOption{ProductDetailID: detail.ID, Name: "color", Value: "red"}
	require.NoError(t, repo.CreateOption(context.Background(), &option))
	variant := domain.ProductVariant{ProductDetailID: detail.ID, ProductPriceID: price.ID, OptionIDs: []string{option.ID}}
	require.NoError(t, repo.CreateVariant(context.Background(), &variant))
	return product, detail, price, option, variant
}

func variantRequest(t *testing.T, method, path string, vars map[string]string, body any, member *domain.Member, handler http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(method, path, bytes.NewReader(encoded))
	req = mux.SetURLVars(req, vars)
	if member != nil {
		req = req.WithContext(api.ContextWithMember(req.Context(), member))
	}
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestVariantHandlers_InventoryUsesVariantID(t *testing.T) {
	defer cleanupProducts(t)
	productRepo := database.NewProductRepositoryPGX(testPool)
	_, _, _, _, variant := createVariantFixture(t, productRepo)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	sessionCache := session.NewSessionCache(time.Hour)
	member := createAuthMember(t, memberRepo, sessionCache)
	handler := composeInventoryHandler(database.NewInventoryRepositoryPGX(testPool))

	w := variantRequest(t, http.MethodPost, "/api/inventories", nil, domain.CreateInventoryRequest{ProductVariantID: variant.ID}, member, handler.CreateInventory)
	assert.Equal(t, http.StatusCreated, w.Code)
	var response domain.InventoryResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
	assert.Equal(t, variant.ID, response.Inventory.ProductVariantID)

	w = variantRequest(t, http.MethodPost, "/api/inventories", nil, domain.CreateInventoryRequest{}, member, handler.CreateInventory)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	w = variantRequest(t, http.MethodPost, "/api/inventories", nil, domain.CreateInventoryRequest{ProductVariantID: "not-a-uuid"}, member, handler.CreateInventory)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventoryAPI_CreateIgnoresUnknownFields(t *testing.T) {
	defer cleanupProducts(t)
	repo := database.NewProductRepositoryPGX(testPool)
	_, _, _, _, variant := createVariantFixture(t, repo)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	sessionCache := session.NewSessionCache(time.Hour)
	cookie := authCookie(t, memberRepo, sessionCache, "admin")

	r := mux.NewRouter()
	api.RegisterInventoryRoutes(
		r,
		usecase.NewInventoryService(database.NewInventoryRepositoryPGX(testPool)),
		usecase.NewMemberService(memberRepo, sessionCache),
		usecase.NewSessionService(sessionCache),
	)

	legacyBody := []byte(`{"product_variant_id":"` + variant.ID + `","product_price_id":"legacy-price-id","status":"銷售中"}`)
	w := doInventoryAPIRequest(t, r, http.MethodPost, "/api/inventories", legacyBody, cookie)
	assert.Equal(t, http.StatusCreated, w.Code)
	var legacyResponse domain.InventoryResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&legacyResponse))
	assert.Equal(t, variant.ID, legacyResponse.Inventory.ProductVariantID)
}

func doInventoryAPIRequest(t *testing.T, r *mux.Router, method, path string, body []byte, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestVariantHandlers_RejectCrossProductMutations(t *testing.T) {
	defer cleanupProducts(t)
	repo := database.NewProductRepositoryPGX(testPool)
	product, _, _, _, _ := createVariantFixture(t, repo)
	_, _, _, option, variant := createVariantFixture(t, repo)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	sessionCache := session.NewSessionCache(time.Hour)
	member := createAuthMember(t, memberRepo, sessionCache)
	handler := composeProductHandler(repo)
	w := variantRequest(t, http.MethodPost, "/api/products/"+product.ID+"/detail/variants", map[string]string{"productId": product.ID}, domain.CreateProductVariantRequest{}, member, handler.CreateVariant)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	updateBody := domain.UpdateProductVariantRequest{ProductPriceID: variant.ProductPriceID, OptionIDs: variant.OptionIDs}
	w = variantRequest(t, http.MethodPost, "/api/products/"+product.ID+"/detail/variants/"+variant.ID+"/update", map[string]string{"productId": product.ID, "variantId": variant.ID}, updateBody, member, handler.UpdateVariant)
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = variantRequest(t, http.MethodPost, "/api/products/"+product.ID+"/detail/variants/"+variant.ID+"/delete", map[string]string{"productId": product.ID, "variantId": variant.ID}, nil, member, handler.DeleteVariant)
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = variantRequest(t, http.MethodPost, "/api/products/"+product.ID+"/detail/options/"+option.ID+"/delete", map[string]string{"productId": product.ID, "optionId": option.ID}, nil, member, handler.DeleteOption)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestVariantHandlers_DuplicateSKUReturnsSafeConflict(t *testing.T) {
	defer cleanupProducts(t)
	repo := database.NewProductRepositoryPGX(testPool)
	product, detail, price, _, first := createVariantFixture(t, repo)
	sku := "SKU-DUPLICATE"
	first.SKU = &sku
	require.NoError(t, repo.UpdateVariant(context.Background(), &first))

	option := domain.ProductOption{ProductDetailID: detail.ID, Name: "size", Value: "M"}
	require.NoError(t, repo.CreateOption(context.Background(), &option))
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	sessionCache := session.NewSessionCache(time.Hour)
	member := createAuthMember(t, memberRepo, sessionCache)
	handler := composeProductHandler(repo)

	w := variantRequest(t, http.MethodPost, "/api/products/"+product.ID+"/detail/variants", map[string]string{"productId": product.ID}, domain.CreateProductVariantRequest{
		ProductPriceID: price.ID,
		SKU:            first.SKU,
		OptionIDs:      []string{option.ID},
	}, member, handler.CreateVariant)
	assert.Equal(t, http.StatusConflict, w.Code)
	var response map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
	assert.Equal(t, "此 SKU 已存在，請使用其他 SKU", response["error"])
	assert.NotContains(t, w.Body.String(), "SQLSTATE")
	assert.NotContains(t, w.Body.String(), "duplicate sku")
}
