package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/mark-adams/inc/api"
	"github.com/mark-adams/inc/backends"
)

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

	handler := handlers.CombinedLoggingHandler(os.Stdout, api.DefaultRouter)
	http.ListenAndServe(":8080", handler)
}
