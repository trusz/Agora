package app

import (
	"goazuread/src/db"
	"goazuread/src/post"

	_ "github.com/mattn/go-sqlite3"
)

func SetupDB(db *db.DB) error {
	post.CreateDBTable(db)

	return nil
}
