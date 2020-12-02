package stores

type Todo struct {
	ID        string
	Title     string
	Completed bool
	Order     int
}

type Store interface {
	Connect() (bool, error)
	Create(todo *Todo) error
	Clear() error
	Get(id string) (*Todo, error)
	Update(id string, todo *Todo) (*Todo, error)
	Delete(id string) error
	List() ([]Todo, error)
}
