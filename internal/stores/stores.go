package stores

type Todo struct {
	ID        string
	Title     string
	Completed bool
	Order     int
}

type Store interface {
	Connect() error
	Create(list string, todo *Todo) error
	Clear(list string) error
	Get(list string, id string) (*Todo, error)
	Update(list string, id string, todo *Todo) (*Todo, error)
	Delete(list string, id string) error
	List(list string) ([]Todo, error)
}
