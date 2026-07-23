package inventory

import (
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/shared"
)

type InventoryReserved struct {
	shared.BaseEvent
	OrderID  string
	Quantity int
}

type InventoryReleased struct {
	shared.BaseEvent
	OrderID  string
	Quantity int
}

type ReservationConfirmed struct {
	shared.BaseEvent
	OrderID  string
	Quantity int
}
