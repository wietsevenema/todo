package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yfuruyama/crzerolog"

	"github.com/wietsevenema/todo/internal/handler"
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

func getURL() ([]string, error) {
	dbURL := os.Getenv("DB")
	split := strings.SplitN(dbURL, "://", 2)
	if len(split) != 2 {
		return nil, fmt.Errorf("Invalid DB env var")
	}
	return split, nil
}

func newSessionStore() *sessions.CookieStore {
	key := os.Getenv("SECRET_SESSION_KEY")
	var bkey []byte
	if key == "" {
		log.Error().Msg("Set env var SECRET_SESSION_KEY to secure sessions")
		bkey = []byte("secret")
	} else {
		bkey = []byte(key)
	}
	return sessions.NewCookieStore(bkey)
}

func main() {
	rootLogger := zerolog.New(os.Stdout)
	middleware := crzerolog.InjectLogger(&rootLogger)

	r := mux.NewRouter()
	r.Use(middleware)

	s := handler.Service{SessionStore: newSessionStore()}

	dbURL, err := getURL()
	if err != nil {
		rootLogger.Warn().Err(err).Msg("Invalid env var DB")
		dbURL = []string{"memory"}
	}

	switch dbURL[0] {
	case "mysql":
		log.Info().Msg("Connecting with MySQL")
		s.Store = stores.NewSQLStore(dbURL[1])
		err := s.Store.Connect()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to MySQL")
		}
	case "redis":
		log.Info().Msg("Connecting with Redis")
		s.Store = stores.NewRedisStore(dbURL[1])
		err := s.Store.Connect()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to Redis")
		}
	default:
		log.Info().Msg("Storing data in memory (not persistent!)")
		s.Store = stores.NewMemory()
	}

	api := r.PathPrefix("/api").Subrouter()
	api.Use(handler.JsonHeader)
	api.Use(s.SessionHandler)
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
