package model

type MemberType string
type Permission string
type OrderStatus string
type PaymentStatus string

const (
	MemberTypeCustomer MemberType = "customer"
	MemberTypeEmployee MemberType = "employee"
	PermissionAdmin    Permission = "admin"

	OrderStatusPending   OrderStatus   = "pending"
	OrderStatusConfirmed OrderStatus   = "confirmed"
	OrderStatusCompleted OrderStatus   = "completed"
	PaymentStatusPaid    PaymentStatus = "paid"

	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)
