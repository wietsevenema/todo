package stores

import (
	"encoding/json"
	"strings"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

type RedisStore struct {
	DBUrl string
	DB    *redis.Client
}

func NewRedisStore(dbURL string) *RedisStore {
	r := RedisStore{dbURL, nil}
	return &r
}

func (r *RedisStore) Connect() (bool, error) {
	if !strings.HasPrefix(r.DBUrl, "redis://") {
		return false, nil
	}
	dbUrl := strings.TrimPrefix(r.DBUrl, "redis://")

	r.DB = redis.NewClient(&redis.Options{
		Addr:     dbUrl,
		Password: "",
		DB:       0,
	})

	_, err := r.DB.Ping().Result()
	if err != nil {
		return true, err
	}

	return true, nil
}

func (r RedisStore) Create(t *Todo) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	t.ID = id.String()

	tb, err := json.Marshal(t)
	if err != nil {
		return err
	}

	err = r.DB.Set(t.ID, tb, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r RedisStore) Delete(id string) error {
	err := r.DB.Del(id).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r RedisStore) Update(id string, newT *Todo) (*Todo, error) {
	tb, err := r.DB.Get(id).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var t Todo
	json.Unmarshal(tb, &t)

	if newT.Title != "" {
		t.Title = newT.Title
	}
	t.Completed = newT.Completed
	t.Order = newT.Order

	tn, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	err = r.DB.Set(t.ID, tn, 0).Err()
	return &t, nil

}

func (r RedisStore) Get(id string) (*Todo, error) {
	tb, err := r.DB.Get(id).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var t Todo
	json.Unmarshal(tb, &t)

	return &t, nil
}

func (r RedisStore) Clear() error {
	_, err := r.DB.FlushDB().Result()
	if err != nil {
		return err
	}
	return nil
}

func (r RedisStore) List() ([]Todo, error) {
	keys, err := r.DB.Keys("*").Result()
	if err != nil {
		return nil, err
	}
	result := []Todo{}
	for _, k := range keys {
		t, err := r.Get(k)
		if err != nil {
			return nil, err
		}
		if t != nil {
			result = append(result, *t)
		}
	}
	return result, nil
}
