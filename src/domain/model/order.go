package model

import (
	"encoding/json"
	"time"
)

type Order struct {
	ID                string `json:"id"`
	OrderNo           string `json:"order_no"`
	CustomerID        string `json:"customer_id"`
	Status            string `json:"status"`
	PaymentStatus     string `json:"payment_status"`
	FulfillmentStatus string `json:"fulfillment_status"`
	Subtotal          string `json:"subtotal"`
	TotalAmount       string `json:"total_amount"`
	// CustomerSnapshot is the immutable JSON object {name, phone, email} supplied at checkout.
	CustomerSnapshot json.RawMessage `json:"customer_snapshot"`
	// ShippingAddressSnapshot is the immutable JSON object {delivery_method, address?}.
	// address is omitted for email delivery. The JSON shape keeps the existing snapshot columns
	// compatible with orders created before checkout customer details were introduced.
	ShippingAddressSnapshot json.RawMessage `json:"shipping_address_snapshot"`
	CreatedAt               time.Time       `json:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at"`
	Items                   []OrderItem     `json:"items"`
}

type OrderItem struct {
	ID              string    `json:"id"`
	OrderID         string    `json:"order_id"`
	ProductPriceID  string    `json:"product_price_id"`
	Quantity        int       `json:"quantity"`
	UnitPrice       string    `json:"unit_price"`
	LineTotal       string    `json:"line_total"`
	ProductSnapshot []byte    `json:"product_snapshot"`
	CreatedAt       time.Time `json:"created_at"`
}

type OrderStatusHistory struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	ChangedBy string    `json:"changed_by"`
	CreatedAt time.Time `json:"created_at"`
}

type Payment struct {
	ID         string    `json:"id"`
	OrderID    string    `json:"order_id"`
	MemberID   string    `json:"member_id"`
	Method     string    `json:"method"`
	Status     string    `json:"status"`
	Amount     string    `json:"amount"`
	MaskedCard string    `json:"masked_card"`
	Last4      string    `json:"last4"`
	CreatedAt  time.Time `json:"created_at"`
}
