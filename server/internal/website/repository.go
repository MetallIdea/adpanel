package website

type Repository interface {
	GetAll() ([]Website, error)
	GetByID(id int64) (*Website, error)
	Create(site *Website) error
	Update(site *Website) error
	Delete(id int64) error
}
