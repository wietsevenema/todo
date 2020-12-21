package stores

import "github.com/google/uuid"

type Memory struct {
	todos map[string]Todo
}

func NewMemory() *Memory {
	m := Memory{}
	m.todos = map[string]Todo{}
	return &m
}

func (m Memory) Create(_ string, todo *Todo) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	todo.ID = id.String()
	m.todos[todo.ID] = *todo
	return nil
}

func (m Memory) Connect() error {
	return nil
}

func (m Memory) Delete(_ string, id string) error {
	delete(m.todos, id)
	return nil
}

func (m Memory) Update(listID string, id string, newT *Todo) (*Todo, error) {
	oldT, err := m.Get(listID, id)
	if err != nil {
		return nil, err
	}
	if oldT != nil {
		if newT.Title != "" {
			oldT.Title = newT.Title
		}
		oldT.Completed = newT.Completed
		oldT.Order = newT.Order
		m.todos[id] = *oldT
		return oldT, nil
	}

	return nil, nil
}

func (m Memory) Get(_ string, id string) (*Todo, error) {
	t := m.todos[id]
	return &t, nil
}

func (m Memory) Clear(_ string) error {
	for k := range m.todos {
		delete(m.todos, k)
	}
	return nil
}

func (m Memory) List(_ string) ([]Todo, error) {
	result := []Todo{}
	for _, t := range m.todos {
		result = append(result, t)
	}
	return result, nil
}
