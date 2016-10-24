package main

import (
	"fmt"
	"log"

	"crypto/rand"
	"encoding/hex"

	"github.com/go-martini/martini"
	_ "github.com/lib/pq"
	"github.com/mark-adams/inc/backends"
	"gopkg.in/alexcesaro/statsd.v2"
)

var app *martini.ClassicMartini

func getRandomID() (string, error) {
	newID := make([]byte, 16)
	_, err := rand.Read(newID)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(newID), nil
}

func init() {

	app = martini.Classic()
	statsd, err := statsd.New()
	if err != nil {
		log.Printf("error initializing metrics: %s", err)
	}

	app.Get("/_healthcheck", func() string {
		return "OK"
	})

	app.Post("/new", func() (int, string) {
		statsd.Increment("inc.api.create_token")

		id, err := getRandomID()
		if err != nil {
			log.Printf("Error: %s", err)
			return 500, "😞 Something bad happened... try again?"
		}
		db, err := backends.GetBackend()
		if err != nil {
			log.Printf("Database error: %s", err)
			return 500, err.Error()
		}

		err = db.CreateToken(id)
		if err != nil {
			log.Printf("Insert error: %s", err)
			return 500, err.Error()
		}
		return 201, id
	})

	app.Put("/(?P<token>[a-zA-Z0-9]{32})", func(params martini.Params) (int, string) {
		statsd.Increment("inc.api.increment_token")

		db, err := backends.GetBackend()
		if err != nil {
			return 500, err.Error()
		}
		count, err := db.IncrementAndGetToken(params["token"])
		if err != nil {
			return 500, err.Error()
		}

		return 200, fmt.Sprintf("%d", count)
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

	app.Run()
}
