//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apphttp "backend/src/adapter/http"
	database "backend/src/adapter/postgres"
	"backend/src/adapter/session"
	"backend/src/usecase"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

const (
	publicFixtureProductID  = "11111111-1111-1111-1111-111111111111"
	publicFixtureDetailID   = "22222222-2222-2222-2222-222222222222"
	publicFixturePriceAID   = "33333333-3333-3333-3333-333333333333"
	publicFixturePriceBID   = "44444444-4444-4444-4444-444444444444"
	publicFixtureCategoryID = "55555555-5555-5555-5555-555555555555"
)

type publicProductFixture struct {
	ProductID string
	DetailID  string
	PriceIDs  []string
}

func seedPublicProductFixture(t *testing.T, pool *pgxpool.Pool) publicProductFixture {
	t.Helper()

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO categories (id, name)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`,
		publicFixtureCategoryID, "Integration Category")
	require.NoError(t, err)

	var rootID string
	require.NoError(t, tx.QueryRow(ctx, `SELECT id::text FROM members WHERE email = 'root@gmail.com'`).Scan(&rootID))
	_, err = tx.Exec(ctx, `
		INSERT INTO products (id, type, name, status, category_id, member_id)
		VALUES ($1, 'product', $2, 'active', $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			category_id = EXCLUDED.category_id`,
		publicFixtureProductID,
		"Integration Product",
		publicFixtureCategoryID,
		rootID)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `
		INSERT INTO product_details (id, product_id, introduction, usage_instructions, return_policy)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			introduction = EXCLUDED.introduction,
			usage_instructions = EXCLUDED.usage_instructions,
			return_policy = EXCLUDED.return_policy`,
		publicFixtureDetailID,
		publicFixtureProductID,
		"Integration introduction",
		"Integration usage",
		"Integration returns")
	require.NoError(t, err)

	for _, price := range []struct {
		id     string
		label  string
		amount float64
		sort   int
	}{
		{id: publicFixturePriceAID, label: "Standard", amount: 125.50, sort: 1},
		{id: publicFixturePriceBID, label: "Premium", amount: 250, sort: 2},
	} {
		_, err = tx.Exec(ctx, `
			INSERT INTO product_prices (id, product_detail_id, label, amount, currency, sort_order)
			VALUES ($1, $2, $3, $4, 'TWD', $5)
			ON CONFLICT (id) DO UPDATE SET
				product_detail_id = EXCLUDED.product_detail_id,
				label = EXCLUDED.label,
				amount = EXCLUDED.amount,
				currency = EXCLUDED.currency,
				sort_order = EXCLUDED.sort_order`,
			price.id,
			publicFixtureDetailID,
			price.label,
			price.amount,
			price.sort)
		require.NoError(t, err)
	}

	require.NoError(t, tx.Commit(ctx))
	return publicProductFixture{
		ProductID: publicFixtureProductID,
		DetailID:  publicFixtureDetailID,
		PriceIDs:  []string{publicFixturePriceAID, publicFixturePriceBID},
	}
}

func publicProductRouter(pool *pgxpool.Pool) http.Handler {
	productRepo := database.NewProductRepositoryPGX(pool)
	memberRepo := database.NewMemberRepositoryPGX(pool)
	codeRepo := database.NewRegistrationCodeRepositoryPGX(pool)
	sessions := session.NewSessionCache(time.Hour)
	memberService := usecase.NewMemberService(memberRepo, sessions, codeRepo)
	productService := usecase.NewProductService(productRepo)
	sessionService := usecase.NewSessionService(sessions)

	router := mux.NewRouter()
	apphttp.RegisterProductRoutes(router, productService, memberService, sessionService)
	return router
}

func getPublicProductJSON(t *testing.T, router http.Handler, path string, target any) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)
	require.NoError(t, json.NewDecoder(response.Body).Decode(target))
}

func TestPublicProductAPI_ReadsDeterministicFixture(t *testing.T) {
	fixture := seedPublicProductFixture(t, testPool)
	require.Equal(t, fixture, seedPublicProductFixture(t, testPool))
	router := publicProductRouter(testPool)

	var products apphttp.ProductListResponse
	getPublicProductJSON(t, router, "/api/products", &products)
	require.Len(t, products.Products, 1)
	require.Equal(t, fixture.ProductID, products.Products[0].ID)
	require.Equal(t, "Integration Product", products.Products[0].Name)
	require.Equal(t, "active", products.Products[0].Status)
	require.Equal(t, publicFixtureCategoryID, products.Products[0].CategoryID)
	require.Equal(t, "Integration Category", products.Products[0].Category)

	var product apphttp.ProductResponse
	getPublicProductJSON(t, router, "/api/products/"+fixture.ProductID, &product)
	require.Equal(t, fixture.ProductID, product.Product.ID)
	require.Equal(t, "Integration Product", product.Product.Name)

	var detail apphttp.DetailResponse
	getPublicProductJSON(t, router, "/api/products/"+fixture.ProductID+"/detail", &detail)
	require.Equal(t, fixture.DetailID, detail.Detail.ID)
	require.Equal(t, "Integration introduction", detail.Detail.Introduction)
	require.Equal(t, "Integration usage", detail.Detail.UsageInstructions)
	require.Equal(t, "Integration returns", detail.Detail.ReturnPolicy)

	var prices apphttp.PriceListResponse
	getPublicProductJSON(t, router, "/api/products/"+fixture.ProductID+"/detail/prices", &prices)
	require.Len(t, prices.Prices, 2)
	require.Equal(t, fixture.PriceIDs, []string{prices.Prices[0].ID, prices.Prices[1].ID})
	require.Equal(t, "Standard", prices.Prices[0].Label)
	require.Equal(t, 125.50, prices.Prices[0].Amount)
	require.Equal(t, "TWD", prices.Prices[0].Currency)
	require.Equal(t, "Premium", prices.Prices[1].Label)
	require.Equal(t, float64(250), prices.Prices[1].Amount)

	var price apphttp.PriceResponse
	getPublicProductJSON(t, router, "/api/products/"+fixture.ProductID+"/detail/prices/"+fixture.PriceIDs[0], &price)
	require.Equal(t, fixture.PriceIDs[0], price.Price.ID)
	require.Equal(t, "Standard", price.Price.Label)
	require.Equal(t, 125.50, price.Price.Amount)
}
