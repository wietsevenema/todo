package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yfuruyama/crzerolog"

	"github.com/wietsevenema/todo/internal/stores"

	"net/http"
)

func port() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func jsonHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// w.Header().Add("access-control-allow-origin", "*")
		// w.Header().Add("access-control-allow-headers", "*")
		// w.Header().Add("access-control-allow-methods", "*")

		w.Header().Add("content-type", "application/json")
		next.ServeHTTP(w, r)
	})
}

type service struct {
	store stores.Store
}

type todoWithURL struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Order     int    `json:"order"`
	URL       string `json:"url"`
}

func addURL(t stores.Todo) todoWithURL {
	tU := todoWithURL{
		Title:     t.Title,
		Completed: t.Completed,
		Order:     t.Order,
		URL:       "/" + t.ID,
		// URL:       "http://localhost:8080/api/todo/" + t.ID,
	}
	return tU
}

func (s *service) List(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	list, err := s.store.List()
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

func (s *service) Clear(w http.ResponseWriter, r *http.Request) {
	s.store.Clear()
	s.List(w, r)
}

func (s *service) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	err := s.store.Delete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *service) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	decoder := json.NewDecoder(r.Body)

	var newT stores.Todo
	err := decoder.Decode(&newT)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := s.store.Update(id, &newT)
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

func (s *service) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	t, err := s.store.Get(vars["id"])
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

func (s *service) Create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t stores.Todo
	err := decoder.Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = s.store.Create(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(addURL(t)) // Can have no id when using auto increment (switch to uuid)
}

func connectStore(dbURL string) {

}

func main() {
	rootLogger := zerolog.New(os.Stdout)
	middleware := crzerolog.InjectLogger(&rootLogger)

	r := mux.NewRouter()
	r.Use(middleware)

	var s service
	dbURL := os.Getenv("DB")

	var m stores.Store

	m = stores.NewSQLStore(dbURL)
	accept, err := m.Connect()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to MySQL")
	}

	if accept != true {
		m = stores.NewRedisStore(dbURL)
		accept, err := m.Connect()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to Redis")
		}

		if !accept {
			log.Info().Msg("Storing data in memory (not persistent)")
			m := stores.NewMemory()
			m.Connect()
		}
	}
	log.Info().Msg(fmt.Sprintf("Connected with %v", reflect.TypeOf(m)))
	s = service{m}

	api := r.PathPrefix("/api").Subrouter()
	api.Use(jsonHeader)
	api.Methods(http.MethodOptions).HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
	api.Path("/todo").Methods(http.MethodPost).HandlerFunc(s.Create)
	api.Path("/todo").Methods(http.MethodGet).HandlerFunc(s.List)
	api.Path("/todo").Methods(http.MethodDelete).HandlerFunc(s.Clear)
	api.Path("/todo/{id}").Methods(http.MethodGet).HandlerFunc(s.Get)
	api.Path("/todo/{id}").Methods(http.MethodDelete).HandlerFunc(s.Delete)
	api.Path("/todo/{id}").Methods(http.MethodPatch, http.MethodPut).HandlerFunc(s.Update)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("/app/dist")))
	log.Info().Msg("Listening on port " + port())
	log.Fatal().Err(http.ListenAndServe(":"+port(), r)).Msg("Can't start service")

}

func sendErr(w http.ResponseWriter, err error) {
	http.Error(w, "Error retrieving products", http.StatusInternalServerError)
	fmt.Println(err) //fixme logger
}
