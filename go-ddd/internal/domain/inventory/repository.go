package inventory

type Repository interface {
	Save(p *Product) error
	FindByID(id string) (*Product, error)
}
