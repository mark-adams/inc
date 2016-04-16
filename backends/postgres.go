package backends

import (
	"database/sql"

	"log"
)

type PostgresBackend struct {
	db *sql.DB
}

func NewPostgresBackend(databaseURL string) (*PostgresBackend, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	return &PostgresBackend{
		db: db,
	}, nil
}

func (p *PostgresBackend) CreateSchema() error {
	_, err := p.db.Exec("CREATE TABLE IF NOT EXISTS counters (id char(32) PRIMARY KEY, count bigint);")
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresBackend) DropSchema() error {
	_, err := p.db.Exec("DROP TABLE IF EXISTS counters;")
	return err
}

func (p *PostgresBackend) CreateToken(token string) error {
	_, err := p.db.Exec("INSERT into counters (id, count) VALUES ($1, 0)", token)
	return err
}

func (p *PostgresBackend) IncrementAndGetToken(token string) (int64, error) {
	tx, err := p.db.Begin()
	if err != nil {
		defer tx.Rollback()
		log.Printf("Error starting transaction: %s", err)
		return 0, errDatabase
	}

	result := tx.QueryRow("SELECT count from counters WHERE id = $1 FOR UPDATE", token)
	var count int64
	err = result.Scan(&count)
	if err == sql.ErrNoRows {
		return 0, errInvalidToken
	}
	if err != nil {
		defer tx.Rollback()
		log.Printf("Error querying database: %s", err)
		return 0, errDatabase
	}

	_, err = tx.Exec("UPDATE counters SET count = count + 1 WHERE id = $1", token)
	if err != nil {
		defer tx.Rollback()
		log.Printf("Error updating counter: %s", err)
		return 0, errDatabase
	}

	tx.Commit()
	return count + 1, nil
}
