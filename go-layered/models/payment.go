package models

type PaymentStatus string

const (
	PaymentPending    PaymentStatus = "pending"
	PaymentAuthorized PaymentStatus = "authorized"
	PaymentCaptured   PaymentStatus = "captured"
	PaymentVoided     PaymentStatus = "voided"
	PaymentRefunded   PaymentStatus = "refunded"
)

type Payment struct {
	ID       string
	OrderID  string
	Amount   int64
	Currency string
	Status   PaymentStatus
	Refunded int64
}
