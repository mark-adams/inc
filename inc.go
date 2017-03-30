package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/mark-adams/inc/backends"
)

var app *mux.Router

func getRandomID() (string, error) {
	newID := make([]byte, 16)
	_, err := rand.Read(newID)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(newID), nil
}

func init() {
	app = mux.NewRouter()

	metrics, err := statsd.New(statsd.Address(os.Getenv("STATSD_HOST")))
	if err != nil && os.Getenv("STATSD_HOST") != "" {
		log.Printf("error initializing metrics: %s", err)
	}

	app.Path("/_healthcheck").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		fmt.Fprint("OK")
	})

	app.Path("/new").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

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
		w.Write([]byte(id))
	})

	app.Path("/{token:[a-zA-Z0-9]{32}}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		metrics.Increment("inc.api.increment_token")
		params := mux.Vars(r)
		db, err := backends.GetBackend()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		count, err := db.IncrementAndGetToken(params["token"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%d", count)
	})

	app.Path("/{token:[a-zA-Z0-9]{32}}/{namespace:.+}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "PUT" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		metrics.Increment("inc.api.increment_namespace_token")
		params := mux.Vars(r)
		db, err := backends.GetBackend()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		count, err := db.IncrementAndGetNamespacedToken(params["token"], params["namespace"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%d", count)

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
