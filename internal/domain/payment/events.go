package payment

import (
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
)

type PaymentAuthorized struct {
	shared.BaseEvent
	OrderID string
	Amount  shared.Money
}

type PaymentCaptured struct {
	shared.BaseEvent
	OrderID string
	Amount  shared.Money
}

type PaymentVoided struct {
	shared.BaseEvent
	OrderID string
	Amount  shared.Money
}

type PaymentRefunded struct {
	shared.BaseEvent
	OrderID string
	Amount  shared.Money
}
