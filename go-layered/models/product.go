package models

type Product struct {
	ID    string
	Name  string
	Stock int
}

type Reservation struct {
	ProductID string
	OrderID   string
	Quantity  int
}
