package handler

import (
	"context"
	"encoding/base32"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
	"github.com/wietsevenema/todo/internal/stores"
)

type sessionKey string

const sessionIDKey sessionKey = "sessionID"

func JsonHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		next.ServeHTTP(w, r)
	})
}

type Service struct {
	Store        stores.Store
	SessionStore sessions.Store
}

func (s *Service) SessionHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.SessionStore.Get(r, "session")
		if session.Values["ID"] == nil {
			session.Values["ID"] = strings.TrimRight(
				base32.StdEncoding.EncodeToString(
					securecookie.GenerateRandomKey(16)), "=")
		}
		ctx := context.WithValue(r.Context(), sessionIDKey, session.Values["ID"])
		err := session.Save(r, w)
		if err != nil {
			log.Err(err).Msg("Error saving session to response")
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type todoWithURL struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Order     int    `json:"order"`
	URL       string `json:"url"`
}

func listID(r *http.Request) string {
	return r.Context().Value(sessionIDKey).(string)
}

func addURL(t stores.Todo) todoWithURL {
	tU := todoWithURL{
		Title:     t.Title,
		Completed: t.Completed,
		Order:     t.Order,
		URL:       "/" + t.ID,
		// URL: "http://localhost:8080/api/todo/" + t.ID,
	}
	return tU
}

func (s *Service) List(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	list, err := s.Store.List(listID(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	result := []todoWithURL{}
	for _, t := range list {
		result = append(result, addURL(t))
	}
	json.NewEncoder(w).Encode(result)
}

func (s *Service) Clear(w http.ResponseWriter, r *http.Request) {
	s.Store.Clear(listID(r))
	s.List(w, r)
}

func (s *Service) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	err := s.Store.Delete(listID(r), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	decoder := json.NewDecoder(r.Body)

	var newT stores.Todo
	err := decoder.Decode(&newT)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := s.Store.Update(listID(r), id, &newT)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if res == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(addURL(*res))
}

func (s *Service) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	t, err := s.Store.Get(listID(r), vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if t != nil {
		json.NewEncoder(w).Encode(addURL(*t))
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (s *Service) Create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t stores.Todo
	err := decoder.Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = s.Store.Create(listID(r), &t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(addURL(t))
}
