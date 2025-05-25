package app

import (
	"goazuread/src/db"

	_ "github.com/mattn/go-sqlite3"
)

func SetupDB(db *db.DB) error {
	// Create the posts table if it doesn't exist
	query := `
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil

}
