package post

import "goazuread/src/db"

type PostHandler struct {
	db *db.DB
}

func NewPostHandler(db *db.DB) *PostHandler {
	return &PostHandler{db: db}
}
