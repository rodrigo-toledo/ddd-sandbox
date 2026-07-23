package order

type Repository interface {
	Save(o *Order) error
	FindByID(id string) (*Order, error)
}
