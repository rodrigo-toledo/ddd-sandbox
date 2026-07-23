package order

import (
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
)

type OrderPlaced struct {
	shared.BaseEvent
	CustomerID string
	Total      shared.Money
	Items      []Item
}

type OrderConfirmed struct {
	shared.BaseEvent
	Total shared.Money
}

type OrderShipped struct {
	shared.BaseEvent
}

type OrderDelivered struct {
	shared.BaseEvent
}

type OrderCancelled struct {
	shared.BaseEvent
}
