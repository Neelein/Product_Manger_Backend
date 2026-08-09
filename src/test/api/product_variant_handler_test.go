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
