package post

import (
	"agora/src/db"
	"agora/src/post/comment"
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
