package post

import (
	"agora/src/comment"
	"agora/src/db"
)

type PostHandler struct {
	db *db.DB
	ch *comment.CommentHandler
}

func NewPostHandler(db *db.DB, ch *comment.CommentHandler) *PostHandler {
	return &PostHandler{
		db: db,
		ch: ch,
	}
}
