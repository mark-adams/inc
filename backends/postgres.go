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

func (p *PostgresBackend) Close() error {
	return p.db.Close()
}

func (p *PostgresBackend) CreateSchema() error {
	_, err := p.db.Exec(`CREATE TABLE IF NOT EXISTS counters
							(id char(32) PRIMARY KEY, count bigint);

						  CREATE TABLE IF NOT EXISTS namespaced_counters
							(id bigserial PRIMARY KEY, token char(32),
							 namespace varchar, count bigint,
							 UNIQUE (token, namespace),
							 FOREIGN KEY (token) REFERENCES counters (ID))`)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresBackend) DropSchema() error {
	_, err := p.db.Exec(`DROP TABLE IF EXISTS counters;
						 DROP TABLE IF EXISTS semantic_tokens;`)
	return err
}

func (p *PostgresBackend) CreateToken(token string) error {
	_, err := p.db.Exec("INSERT into counters (id, count) VALUES ($1, 0)", token)
	return err
}

func (p *PostgresBackend) IncrementAndGetNamespacedToken(token string, namespace string) (int64, error) {
	tx, err := p.db.Begin()
	if err != nil {
		defer tx.Rollback()
		log.Printf("Error starting transaction: %s", err)
		return 0, errDatabase
	}

	var count int64

	err = tx.QueryRow(`SELECT s.count
						FROM namespaced_counters AS s
						WHERE token = $1 AND namespace = $2 FOR UPDATE`, token, namespace).Scan(&count)

	// No counter found for this namespace, create namespace and initialize count to 0
	if err == sql.ErrNoRows {
		_, err := p.db.Exec(`INSERT INTO namespaced_counters (token, namespace, count)
							 VALUES ($1, $2, 0)`, token, namespace)
		if err != nil {
			defer tx.Rollback()
			log.Printf("Error creating namespaced counter: %s", err)
			return 0, errDatabase
		}

		tx.Commit()

		return 0, nil
	}
	count += 1
	_, err = tx.Exec(`UPDATE namespaced_counters
					  SET count = $1
					  WHERE token = $2 AND namespace = $3`, count, token, namespace)
	tx.Commit()

	if err != nil {
		defer tx.Rollback()
		log.Printf("Error updating namespaced counter: %s", err)
		return 0, errDatabase
	}

	return count, nil
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
