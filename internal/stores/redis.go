package stores

import (
	"encoding/json"
	"fmt"

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

func (r *RedisStore) Connect() error {
	r.DB = redis.NewClient(&redis.Options{
		Addr:     r.DBUrl,
		Password: "",
		DB:       0,
	})

	_, err := r.DB.Ping().Result()
	if err != nil {
		return err
	}

	return nil
}

func (r RedisStore) Create(listID string, t *Todo) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	t.ID = id.String()

	tb, err := json.Marshal(t)
	if err != nil {
		return err
	}

	err = r.DB.Set(fmt.Sprintf("%s-%s", listID, t.ID), tb, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r RedisStore) Delete(listID string, id string) error {
	err := r.DB.Del(fmt.Sprintf("%s-%s", listID, id)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r RedisStore) Update(listID string, id string, newT *Todo) (*Todo, error) {
	tb, err := r.DB.Get(fmt.Sprintf("%s-%s", listID, id)).Bytes()
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

	err = r.DB.Set(fmt.Sprintf("%s-%s", listID, t.ID), tn, 0).Err()
	return &t, nil

}

func (r RedisStore) Get(listID string, id string) (*Todo, error) {
	tb, err := r.DB.Get(fmt.Sprintf("%s-%s", listID, id)).Bytes()
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

func (r RedisStore) Clear(listID string) error {
	keys, err := r.DB.Keys(fmt.Sprintf("%s-*", listID)).Result()
	if err != nil {
		return err
	}

	for _, k := range keys {
		r.DB.Del(k).Err()
	}
	return nil
}

func (r RedisStore) List(listID string) ([]Todo, error) {
	keys, err := r.DB.Keys(fmt.Sprintf("%s-*", listID)).Result()
	if err != nil {
		return nil, err
	}
	result := []Todo{}
	for _, k := range keys {
		tb, err := r.DB.Get(k).Bytes()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return nil, err
		}
		var t Todo
		json.Unmarshal(tb, &t)
		result = append(result, t)
	}
	return result, nil
}
