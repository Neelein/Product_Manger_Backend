package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"backend/src/domain/model"
	"backend/src/usecase"
	"github.com/gorilla/mux"
)

type OrderHandler struct{ service usecase.OrderService }

func NewOrderHandler(service usecase.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}
func orderError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	msg := err.Error()
	switch {
	case errors.Is(err, model.ErrForbidden), errors.Is(err, model.ErrOrderForbidden):
		status = http.StatusForbidden
	case errors.Is(err, model.ErrOrderNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrInvalidOrder), errors.Is(err, model.ErrInvalidOrderTransition), errors.Is(err, model.ErrOrderNotCancellable), errors.Is(err, model.ErrInsufficientInventory):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrPriceNotFound):
		status = http.StatusNotFound
	}
	writeError(w, status, msg)
}
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	items := make([]usecase.OrderCreateItem, len(req.Items))
	for i, v := range req.Items {
		items[i] = usecase.OrderCreateItem{ProductPriceID: v.ProductPriceID, Quantity: v.Quantity}
	}
	shippingAddress := ""
	if req.ShippingAddress != nil {
		shippingAddress = req.ShippingAddress.Address
	}
	o, err := h.service.Create(r.Context(), MemberFromContext(r.Context()), usecase.OrderCreateInput{
		Items:           items,
		Customer:        usecase.OrderCustomerInput{Name: req.Customer.Name, Phone: req.Customer.Phone, Email: req.Customer.Email},
		DeliveryMethod:  req.DeliveryMethod,
		ShippingAddress: shippingAddress,
	})
	if err != nil {
		orderError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, OrderResponse{Order: orderResponseDTO(*o)})
}
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("page_size"))
	orders, total, err := h.service.List(r.Context(), MemberFromContext(r.Context()), q.Get("status"), page, size)
	if err != nil {
		orderError(w, err)
		return
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	responseOrders := make([]OrderResponseDTO, len(orders))
	for i, order := range orders {
		responseOrders[i] = orderResponseDTO(order)
	}
	writeJSON(w, 200, OrderListResponse{Orders: responseOrders, Total: total, Page: page, PageSize: size})
}
func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	o, err := h.service.GetByID(r.Context(), MemberFromContext(r.Context()), mux.Vars(r)["orderId"])
	if err != nil {
		orderError(w, err)
		return
	}
	writeJSON(w, 200, OrderResponse{Order: orderResponseDTO(*o)})
}
func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Cancel(r.Context(), MemberFromContext(r.Context()), mux.Vars(r)["orderId"]); err != nil {
		orderError(w, err)
		return
	}
	writeJSON(w, 200, map[string]string{"message": "order cancelled"})
}
func (h *OrderHandler) Status(w http.ResponseWriter, r *http.Request) {
	var req UpdateOrderStatusRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	o, err := h.service.UpdateStatus(r.Context(), MemberFromContext(r.Context()), mux.Vars(r)["orderId"], req.Status)
	if err != nil {
		orderError(w, err)
		return
	}
	writeJSON(w, 200, OrderResponse{Order: orderResponseDTO(*o)})
}

func orderResponseDTO(order model.Order) OrderResponseDTO {
	items := make([]OrderItemResponseDTO, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemResponseDTO{
			ID:              item.ID,
			OrderID:         item.OrderID,
			ProductPriceID:  item.ProductPriceID,
			Quantity:        item.Quantity,
			UnitPrice:       item.UnitPrice,
			LineTotal:       item.LineTotal,
			ProductSnapshot: objectSnapshot(item.ProductSnapshot),
			CreatedAt:       item.CreatedAt,
		}
	}
	return OrderResponseDTO{
		ID:                      order.ID,
		OrderNo:                 order.OrderNo,
		CustomerID:              order.CustomerID,
		Status:                  order.Status,
		PaymentStatus:           order.PaymentStatus,
		FulfillmentStatus:       order.FulfillmentStatus,
		Subtotal:                order.Subtotal,
		TotalAmount:             order.TotalAmount,
		CustomerSnapshot:        objectSnapshot(order.CustomerSnapshot),
		ShippingAddressSnapshot: objectSnapshot(order.ShippingAddressSnapshot),
		CreatedAt:               order.CreatedAt,
		UpdatedAt:               order.UpdatedAt,
		Items:                   items,
	}
}

func objectSnapshot(value []byte) json.RawMessage {
	if len(value) > 0 && json.Valid(value) && value[0] == '{' {
		return json.RawMessage(append([]byte(nil), value...))
	}
	return json.RawMessage(`{}`)
}
func (h *OrderHandler) History(w http.ResponseWriter, r *http.Request) {
	history, err := h.service.History(r.Context(), MemberFromContext(r.Context()), mux.Vars(r)["orderId"])
	if err != nil {
		orderError(w, err)
		return
	}
	writeJSON(w, 200, OrderHistoryResponse{History: history})
}
