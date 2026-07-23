package models

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderConfirmed OrderStatus = "confirmed"
	OrderShipped   OrderStatus = "shipped"
	OrderDelivered OrderStatus = "delivered"
	OrderCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID         string
	CustomerID string
	Status     OrderStatus
	Total      int64
	Currency   string
	PlacedAt   string
	Items      []OrderItem
}

type OrderItem struct {
	ProductID string
	Quantity  int
	UnitPrice int64
	Currency  string
}
