package main

import (
	"fmt"
	"log"

	"crypto/rand"
	"database/sql"
	"encoding/hex"

	"github.com/go-martini/martini"
)

const errDatabase = "😧 The database is having some trouble... try again?"

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
	app.Post("/new", func() (int, string) {
		id, err := getRandomID()
		if err != nil {
			log.Printf("Error: %s", err)
			return 500, "😞 Something bad happened... try again?"
		}
		db, err := GetDatabase()
		if err != nil {
			log.Printf("Database error: %s", err)
			return 500, errDatabase
		}

		_, err = db.Exec("INSERT into counters (id, count) VALUES ($1, 0)", id)
		if err != nil {
			log.Printf("Insert error: %s", err)
			return 500, errDatabase
		}
		return 201, id
	})

	app.Put("/(?P<token>[a-zA-Z0-9]{32})", func(params martini.Params) (int, string) {
		db, err := GetDatabase()
		if err != nil {
			return 500, errDatabase
		}
		defer db.Close()

		tx, err := db.Begin()
		if err != nil {
			defer tx.Rollback()
			return 500, errDatabase
		}

		result := tx.QueryRow("SELECT count from counters WHERE id = $1 FOR UPDATE", params["token"])
		var count uint64
		err = result.Scan(&count)
		if err == sql.ErrNoRows {
			return 404, "404 page not found"
		}
		if err != nil {
			defer tx.Rollback()
			log.Printf("%s", err)
			return 500, errDatabase
		}

		_, err = tx.Exec("UPDATE counters SET count = count + 1 WHERE id = $1", params["token"])
		if err != nil {
			defer tx.Rollback()
			log.Printf("%s", err)
			return 500, errDatabase
		}

		tx.Commit()
		return 200, fmt.Sprintf("%d", count+1)
	})
}

func main() {
	db, err := GetDatabase()
	if err != nil {
		log.Fatal(err)
	}

	err = CreateDatabaseSchema(db)
	if err != nil {
		log.Fatal(err)
	}

	app.Run()
}
