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
	database "backend/src/adapter/postgres"
	"backend/src/adapter/session"
	"backend/src/domain/model"
	"backend/src/usecase"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type orderAPIFixture struct {
	router                 http.Handler
	customer, other, staff *model.Member
	customerSession        *model.Session
	otherSession           *model.Session
	staffSession           *model.Session
	priceID                string
	secondPriceID          string
	productID              string
	sessions               *session.SessionCache
}

func newOrderAPIFixture(t *testing.T, firstStock, secondStock int) *orderAPIFixture {
	t.Helper()
	ctx := context.Background()
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	productRepo := database.NewProductRepositoryPGX(testPool)
	inventoryRepo := database.NewInventoryRepositoryPGX(testPool)
	orderRepo := database.NewOrderRepositoryPGX(testPool)
	paymentRepo := database.NewPaymentRepository(testPool)
	sessions := session.NewSessionCache(time.Hour)

	createMember := func(name, memberType string) *model.Member {
		member := &model.Member{Email: "order-api-" + name + "-" + t.Name() + "@example.com", Name: name, MemberType: memberType}
		_, err := testPool.Exec(ctx, `INSERT INTO members(email,password,name,member_type,permission) VALUES($1,'pw',$2,$3,$4)`, member.Email, member.Name, member.MemberType, memberType)
		require.NoError(t, err)
		require.NoError(t, testPool.QueryRow(ctx, `SELECT id,created_at,updated_at FROM members WHERE email=$1`, member.Email).Scan(&member.ID, &member.CreatedAt, &member.UpdatedAt))
		return member
	}
	customer := createMember("Customer", "customer")
	other := createMember("Other", "customer")
	staff := createMember("Staff", "employee")

	newSession := func(memberID string) *model.Session {
		s := &model.Session{MemberID: memberID}
		require.NoError(t, sessions.Create(ctx, s))
		return s
	}

	product := model.Product{Name: "Order API Product " + t.Name()}
	require.NoError(t, productRepo.Create(ctx, &product))
	image := model.ProductImage{ProductID: product.ID, Filename: "catalog-image.jpg"}
	require.NoError(t, productRepo.CreateImage(ctx, &image))
	detail := model.ProductDetail{ProductID: product.ID}
	require.NoError(t, productRepo.CreateDetail(ctx, &detail))
	price := model.ProductPrice{ProductDetailID: detail.ID, Label: "API price", Amount: 125, Currency: "TWD"}
	require.NoError(t, productRepo.CreatePrice(ctx, &price))
	secondPrice := model.ProductPrice{ProductDetailID: detail.ID, Label: "API second price", Amount: 80, Currency: "TWD"}
	require.NoError(t, productRepo.CreatePrice(ctx, &secondPrice))
	createInventory := func(priceID string, quantity int) {
		inventory := model.Inventory{ProductPriceID: priceID, Status: "銷售中"}
		require.NoError(t, inventoryRepo.CreateInventory(ctx, &inventory))
		for i := 0; i < quantity; i++ {
			require.NoError(t, inventoryRepo.CreateItem(ctx, &model.InventoryItem{InventoryID: inventory.ID, ItemCode: priceID + "-" + string(rune('a'+i)), Status: "可用"}))
		}
	}
	createInventory(price.ID, firstStock)
	createInventory(secondPrice.ID, secondStock)

	router := mux.NewRouter()
	memberService := usecase.NewMemberService(memberRepo, sessions)
	api.RegisterOrderRoutes(router, usecase.NewOrderService(orderRepo), memberService, usecase.NewSessionService(sessions))
	api.RegisterPaymentRoutes(router, usecase.NewPaymentService(paymentRepo), memberService, usecase.NewSessionService(sessions))

	f := &orderAPIFixture{router: router, customer: customer, other: other, staff: staff, customerSession: newSession(customer.ID), otherSession: newSession(other.ID), staffSession: newSession(staff.ID), priceID: price.ID, secondPriceID: secondPrice.ID, productID: product.ID, sessions: sessions}
	t.Cleanup(func() {
		sessions.Stop()
		_, _ = testPool.Exec(ctx, `DELETE FROM payments WHERE order_id IN (SELECT id FROM orders WHERE customer_id = ANY($1::uuid[]))`, []string{customer.ID, other.ID, staff.ID})
		_, _ = testPool.Exec(ctx, `DELETE FROM orders WHERE customer_id = ANY($1::uuid[])`, []string{customer.ID, other.ID, staff.ID})
		_, _ = testPool.Exec(ctx, `DELETE FROM products WHERE id=$1`, product.ID)
		_, _ = testPool.Exec(ctx, `DELETE FROM members WHERE id = ANY($1::uuid[])`, []string{customer.ID, other.ID, staff.ID})
	})
	return f
}

func TestPaymentAPIValidatesOwnershipDuplicatePaymentAndConfirmation(t *testing.T) {
	f := newOrderAPIFixture(t, 3, 3)
	created := f.request(t, f.customerSession, http.MethodPost, "/api/orders", orderBody(f.priceID, 1))
	require.Equal(t, http.StatusCreated, created.Code)
	order := decodeOrder(t, created)
	paymentBody := map[string]string{"method": "CREDIT_CARD", "card_number": "4242424242424242", "cvv": "123", "otp": "1234567"}
	require.Equal(t, http.StatusForbidden, f.request(t, f.otherSession, http.MethodPost, "/api/orders/"+order.ID+"/payments", paymentBody).Code)
	require.Equal(t, http.StatusBadRequest, f.request(t, f.customerSession, http.MethodPost, "/api/orders/"+order.ID+"/payments", map[string]string{"method": "credit_card", "card_number": "4242424242424243", "cvv": "123", "otp": "1234567"}).Code)
	require.Equal(t, http.StatusCreated, f.request(t, f.customerSession, http.MethodPost, "/api/orders/"+order.ID+"/payments", paymentBody).Code)
	require.Equal(t, http.StatusConflict, f.request(t, f.customerSession, http.MethodPost, "/api/orders/"+order.ID+"/payments", paymentBody).Code)
	var status, paymentStatus string
	require.NoError(t, testPool.QueryRow(context.Background(), `SELECT status,payment_status FROM orders WHERE id=$1`, order.ID).Scan(&status, &paymentStatus))
	require.Equal(t, "confirmed", status)
	require.Equal(t, "paid", paymentStatus)
	var cardCount int
	require.NoError(t, testPool.QueryRow(context.Background(), `SELECT count(*) FROM payments WHERE order_id=$1 AND masked_card='**** **** **** 4242'`, order.ID).Scan(&cardCount))
	require.Equal(t, 1, cardCount)
	var method string
	require.NoError(t, testPool.QueryRow(context.Background(), `SELECT method FROM payments WHERE order_id=$1`, order.ID).Scan(&method))
	require.Equal(t, "credit_card", method)
}

func TestPaymentExpirationReleasesReservationsAndIsIdempotent(t *testing.T) {
	f := newOrderAPIFixture(t, 2, 2)
	created := f.request(t, f.customerSession, http.MethodPost, "/api/orders", orderBody(f.priceID, 1))
	order := decodeOrder(t, created)
	old := time.Now().UTC().Add(-11 * time.Minute)
	_, err := testPool.Exec(context.Background(), `UPDATE orders SET created_at=$2 WHERE id=$1`, order.ID, old)
	require.NoError(t, err)
	repo := database.NewPaymentRepository(testPool)
	count, err := repo.ExpirePending(context.Background(), time.Now().UTC().Add(-10*time.Minute), time.Now().UTC())
	require.NoError(t, err)
	require.Equal(t, 1, count)
	count, err = repo.ExpirePending(context.Background(), time.Now().UTC().Add(-10*time.Minute), time.Now().UTC())
	require.NoError(t, err)
	require.Zero(t, count)
	var status, paymentStatus, reservationStatus string
	require.NoError(t, testPool.QueryRow(context.Background(), `SELECT status,payment_status FROM orders WHERE id=$1`, order.ID).Scan(&status, &paymentStatus))
	require.NoError(t, testPool.QueryRow(context.Background(), `SELECT status FROM inventory_reservations WHERE order_id=$1 LIMIT 1`, order.ID).Scan(&reservationStatus))
	require.Equal(t, "cancelled", status)
	require.Equal(t, "failed", paymentStatus)
	require.Equal(t, "released", reservationStatus)
}

func (f *orderAPIFixture) request(t *testing.T, memberSession *model.Session, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		require.NoError(t, err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_key", Value: memberSession.SessionKey})
	rr := httptest.NewRecorder()
	f.router.ServeHTTP(rr, req)
	return rr
}

func orderBody(priceID string, quantity int) map[string]any {
	return map[string]any{"items": []map[string]any{{"product_price_id": priceID, "quantity": quantity, "unit_price": 0.01, "line_total": 0.01}}, "customer": map[string]string{"name": "Checkout Customer", "phone": "0912345678", "email": "checkout@example.com"}, "delivery_method": "home_address", "shipping_address": map[string]string{"address": "Taipei"}}
}

func decodeOrder(t *testing.T, rr *httptest.ResponseRecorder) api.OrderResponseDTO {
	t.Helper()
	var response api.OrderResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&response))
	return response.Order
}

func TestOrderAPIAcceptanceAuthorizationPricingReservationAndTransitions(t *testing.T) {
	f := newOrderAPIFixture(t, 8, 8)

	customerCreate := f.request(t, f.customerSession, http.MethodPost, "/api/orders", orderBody(f.priceID, 2))
	require.Equal(t, http.StatusCreated, customerCreate.Code)
	customerOrder := decodeOrder(t, customerCreate)
	require.Equal(t, "250.00", customerOrder.TotalAmount)
	require.Equal(t, "125.00", customerOrder.Items[0].UnitPrice)
	var customerSnapshot map[string]any
	require.NoError(t, json.Unmarshal(customerOrder.CustomerSnapshot, &customerSnapshot))
	require.Equal(t, "Checkout Customer", customerSnapshot["name"])
	var shippingSnapshot map[string]any
	require.NoError(t, json.Unmarshal(customerOrder.ShippingAddressSnapshot, &shippingSnapshot))
	require.Equal(t, "Taipei", shippingSnapshot["address"])
	var productSnapshot map[string]any
	require.NoError(t, json.Unmarshal(customerOrder.Items[0].ProductSnapshot, &productSnapshot))
	require.Equal(t, "Order API Product "+t.Name(), productSnapshot["product_name"])
	require.Equal(t, "/media/images/products/"+f.productID+"/catalog-image.jpg", productSnapshot["image_url"])

	get := f.request(t, f.customerSession, http.MethodGet, "/api/orders/"+customerOrder.ID, nil)
	require.Equal(t, http.StatusOK, get.Code)
	got := decodeOrder(t, get)
	var retrievedSnapshot map[string]any
	require.NoError(t, json.Unmarshal(got.Items[0].ProductSnapshot, &retrievedSnapshot))
	require.Equal(t, productSnapshot["image_url"], retrievedSnapshot["image_url"])
	var dbAmount float64
	require.NoError(t, testPool.QueryRow(context.Background(), `SELECT unit_price FROM order_items WHERE order_id=$1`, customerOrder.ID).Scan(&dbAmount))
	require.Equal(t, 125.0, dbAmount)
	var activeReservations int
	require.NoError(t, testPool.QueryRow(context.Background(), `SELECT count(*) FROM inventory_reservations WHERE order_id=$1 AND status='active'`, customerOrder.ID).Scan(&activeReservations))
	require.Equal(t, 2, activeReservations)

	staffCreate := f.request(t, f.staffSession, http.MethodPost, "/api/orders", orderBody(f.priceID, 1))
	require.Equal(t, http.StatusCreated, staffCreate.Code)
	staffOrder := decodeOrder(t, staffCreate)
	require.Equal(t, f.staff.ID, staffOrder.CustomerID)

	otherCreate := f.request(t, f.otherSession, http.MethodPost, "/api/orders", orderBody(f.priceID, 1))
	require.Equal(t, http.StatusCreated, otherCreate.Code)
	otherOrder := decodeOrder(t, otherCreate)

	require.Equal(t, http.StatusNotFound, f.request(t, f.customerSession, http.MethodGet, "/api/orders/"+otherOrder.ID, nil).Code)
	list := f.request(t, f.customerSession, http.MethodGet, "/api/orders?page=1&page_size=1&status=pending", nil)
	require.Equal(t, http.StatusOK, list.Code)
	var listResponse api.OrderListResponse
	require.NoError(t, json.NewDecoder(list.Body).Decode(&listResponse))
	require.Equal(t, 1, listResponse.PageSize)
	require.Equal(t, 1, listResponse.Total)
	require.Len(t, listResponse.Orders, 1)
	require.Equal(t, customerOrder.ID, listResponse.Orders[0].ID)

	staffList := f.request(t, f.staffSession, http.MethodGet, "/api/orders?page=1&page_size=2", nil)
	require.Equal(t, http.StatusOK, staffList.Code)
	var staffListResponse api.OrderListResponse
	require.NoError(t, json.NewDecoder(staffList.Body).Decode(&staffListResponse))
	require.Equal(t, 3, staffListResponse.Total)

	require.Equal(t, http.StatusOK, f.request(t, f.customerSession, http.MethodPost, "/api/orders/"+customerOrder.ID+"/cancel", nil).Code)
	var reservationStatus string
	require.NoError(t, testPool.QueryRow(context.Background(), `SELECT status FROM inventory_reservations WHERE order_id=$1 LIMIT 1`, customerOrder.ID).Scan(&reservationStatus))
	require.Equal(t, "released", reservationStatus)

	confirmedCreate := f.request(t, f.customerSession, http.MethodPost, "/api/orders", orderBody(f.priceID, 1))
	require.Equal(t, http.StatusCreated, confirmedCreate.Code)
	confirmedOrder := decodeOrder(t, confirmedCreate)
	require.Equal(t, http.StatusOK, f.request(t, f.staffSession, http.MethodPost, "/api/orders/"+confirmedOrder.ID+"/status", map[string]string{"status": "confirmed"}).Code)
	require.Equal(t, http.StatusBadRequest, f.request(t, f.customerSession, http.MethodPost, "/api/orders/"+confirmedOrder.ID+"/cancel", nil).Code)
	require.Equal(t, http.StatusOK, f.request(t, f.staffSession, http.MethodPost, "/api/orders/"+confirmedOrder.ID+"/status", map[string]string{"status": "completed"}).Code)
	require.Equal(t, http.StatusBadRequest, f.request(t, f.staffSession, http.MethodPost, "/api/orders/"+confirmedOrder.ID+"/status", map[string]string{"status": "confirmed"}).Code)

	history := f.request(t, f.staffSession, http.MethodGet, "/api/orders/"+confirmedOrder.ID+"/history", nil)
	require.Equal(t, http.StatusOK, history.Code)
	var historyResponse api.OrderHistoryResponse
	require.NoError(t, json.NewDecoder(history.Body).Decode(&historyResponse))
	require.Equal(t, []string{"pending", "confirmed", "completed"}, []string{historyResponse.History[0].Status, historyResponse.History[1].Status, historyResponse.History[2].Status})
}

func TestOrderAPITransactionRollbackRemovesOrderItemsAndReservations(t *testing.T) {
	f := newOrderAPIFixture(t, 1, 1)
	rr := f.request(t, f.customerSession, http.MethodPost, "/api/orders", map[string]any{"items": []map[string]any{
		{"product_price_id": f.priceID, "quantity": 1},
		{"product_price_id": f.secondPriceID, "quantity": 2},
	}})
	require.Equal(t, http.StatusBadRequest, rr.Code)

	ctx := context.Background()
	var orders int
	require.NoError(t, testPool.QueryRow(ctx, `SELECT count(*) FROM orders WHERE customer_id=$1`, f.customer.ID).Scan(&orders))
	require.Zero(t, orders)
	for _, table := range []string{"order_items", "inventory_reservations", "order_status_history"} {
		var count int
		require.NoError(t, testPool.QueryRow(ctx, "SELECT count(*) FROM "+table+" WHERE order_id IN (SELECT id FROM orders WHERE customer_id=$1)", f.customer.ID).Scan(&count))
		require.Zero(t, count, table)
	}
}
