package post

import "goazuread/src/db"

const TABLE_QUERY = `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

func CreateDBTable(db *db.DB) error {
	// Create the posts table if it doesn't exist
	_, err := db.Exec(TABLE_QUERY)
	if err != nil {
		return err
	}
	return nil
}
