package stores

import (
	"database/sql"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type SQLStore struct {
	DBUrl string
	DB    *sql.DB
}

func NewSQLStore(dbURL string) *SQLStore {
	s := SQLStore{dbURL, nil}
	return &s
}

func (s *SQLStore) Connect() error {
	db, err := sql.Open("mysql", s.DBUrl)

	if err != nil {
		return err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(10)

	s.DB = db

	err = db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (s SQLStore) Create(_ string, t *Todo) error {
	stmt, err := s.DB.Prepare(`INSERT INTO todos(title, completed, sortOrder) VALUES (
		?,
		?, 
		?
	)`)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(t.Title, t.Completed, t.Order)
	if err != nil {
		return err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = strconv.FormatInt(lastID, 10)
	return nil
}

func (s SQLStore) Delete(_ string, id string) error {
	stmt, err := s.DB.Prepare("DELETE FROM todos WHERE id=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

func (s SQLStore) Update(listID string, id string, newT *Todo) (*Todo, error) {
	t, err := s.Get(listID, id)
	if err != nil {
		return nil, err
	}
	if t != nil {
		if newT.Title != "" {
			t.Title = newT.Title
		}
		t.Completed = newT.Completed
		t.Order = newT.Order

		stmt, err := s.DB.Prepare(`UPDATE todos SET 
			title = ?, 
			completed = ?, 
			sortOrder = ?
			WHERE id=?`)

		if err != nil {
			return nil, err
		}
		_, err = stmt.Exec(t.Title, t.Completed, t.Order, id)
		if err != nil {
			return nil, err
		}

		return t, nil
	}
	return nil, nil
}

func (s SQLStore) Get(_ string, id string) (*Todo, error) {
	rows, err := s.DB.Query("select id, title, completed, sortOrder from todos where id=?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	t := Todo{}
	for rows.Next() {
		err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.Order)
		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s SQLStore) Clear(_ string) error {
	stmt, err := s.DB.Prepare("DELETE FROM todos")
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}

func (s SQLStore) List(_ string) ([]Todo, error) {
	rows, err := s.DB.Query("select id, title, completed, sortOrder from todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Todo{}
	for rows.Next() {
		t := Todo{}
		err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.Order)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}
