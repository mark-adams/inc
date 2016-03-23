package main

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

func getDatabase() (*sql.DB, error) {
	databaseURL := os.Getenv("PG_DB_URL")
	return sql.Open("postgres", databaseURL)
}

func createDatabaseSchema(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS counters (id char(32) PRIMARY KEY, count bigint);")
	if err != nil {
		return err
	}

	return nil
}
