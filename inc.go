package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"

	statsd "gopkg.in/alexcesaro/statsd.v2"

	"github.com/gorilla/handlers"
	_ "github.com/lib/pq"
	"github.com/mark-adams/inc/backends"
	"github.com/pressly/chi"
)

var app *chi.Mux
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

	count, err := db.IncrementAndGetNamespacedToken(chi.URLParam(r, "token"), chi.URLParam(r, "namespae"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%d", count)
}

func init() {
	var err error

	app = chi.NewRouter()
	metrics = &NullMetricsCollector{}
	metrics, err = statsd.New(statsd.Address(os.Getenv("STATSD_HOST")))

	if err != nil && os.Getenv("STATSD_HOST") != "" {
		log.Printf("error initializing metrics: %s", err)
	}

	app.Get("/_healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	app.Post("/new", NewToken)
	app.Route("/:token", func(r chi.Router) {
		r.Put("/", IncrementToken)
		r.Put("/:namespace", IncrementTokenNamespace)
	})
}

func main() {
	db, err := backends.GetBackend()
	if err != nil {
		log.Fatal(err)
	}

	// Create the database schema if needed
	err = db.CreateSchema()
	if err != nil {
		log.Fatal(err)
	}
	db.Close()

	handler := handlers.CombinedLoggingHandler(os.Stdout, app)
	http.ListenAndServe(":8080", handler)
}
