package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DB struct {
	DSN string
	db  *sql.DB
}

func NewDB(dsn string) *DB {
	db := &DB{
		DSN: dsn,
	}
	return db
}

func (db *DB) Open() (err error) {
	// Ensure a DSN is set before attempting to open the database.
	if db.DSN == "" {
		return fmt.Errorf("dsn required")
	}

	// Connect to the database.
	if db.db, err = sql.Open("postgres", db.DSN); err != nil {
		return err
	}

	if err = db.db.Ping(); err != nil {
		return err
	}

	return nil
}

func (db *DB) Close() error {
	// Close database.
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}
