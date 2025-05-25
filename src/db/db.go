package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

// Open creates new connection pool
func Open(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db DB) Close() error {
	if db.DB == nil {
		return nil
	}
	return db.DB.Close()
}
