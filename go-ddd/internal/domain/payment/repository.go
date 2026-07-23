package payment

type Repository interface {
	Save(p *Payment) error
	FindByID(id string) (*Payment, error)
	FindByOrderID(orderID string) (*Payment, error)
}
