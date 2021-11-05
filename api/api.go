package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"

	statsd "gopkg.in/alexcesaro/statsd.v2"

	"github.com/mark-adams/inc/backends"
	"github.com/pressly/chi"
)

// DefaultRouter is a mux populated with the default set of handlers for the inc app
var DefaultRouter *chi.Mux
var metrics MetricCollector

func getRandomID() (string, error) {
	newID := make([]byte, 16)
	_, err := rand.Read(newID)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(newID), nil
}

// NewToken creates a new counting token
func NewToken(w http.ResponseWriter, r *http.Request) {
	metrics.Increment("inc.api.create_token")

	id, err := getRandomID()
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, "😞 Something bad happened... try again?", http.StatusInternalServerError)
		return
	}
	db, err := backends.GetBackend()
	if err != nil {
		log.Printf("Database error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = db.CreateToken(id)
	if err != nil {
		log.Printf("Insert error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, id)
}

// IncrementToken increments an existing token
func IncrementToken(w http.ResponseWriter, r *http.Request) {
	metrics.Increment("inc.api.increment_token")
	db, err := backends.GetBackend()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	count, err := db.IncrementAndGetToken(chi.URLParam(r, "token"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%d", count)
}

// IncrementTokenNamespace increments a specific namespace inside a specific token's context
func IncrementTokenNamespace(w http.ResponseWriter, r *http.Request) {
	metrics.Increment("inc.api.increment_namespace_token")
	db, err := backends.GetBackend()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	count, err := db.IncrementAndGetNamespacedToken(chi.URLParam(r, "token"), chi.URLParam(r, "namespace"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%d", count)
}

func init() {
	var err error

	router := chi.NewRouter()

	metrics = &NullMetricsCollector{}
	metrics, err = statsd.New(statsd.Address(os.Getenv("STATSD_HOST")))
	if err != nil && os.Getenv("STATSD_HOST") != "" {
		log.Printf("error initializing metrics: %s", err)
		metrics = &NullMetricsCollector{}
	}

	router.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	router.Post("/new", NewToken)
	router.Route("/{token}", func(r chi.Router) {
		r.Put("/", IncrementToken)
		r.Put("/{namespace}", IncrementTokenNamespace)
	})
	router.Post("/new", NewToken)

	DefaultRouter = router
}
