package comment

import (
	"agora/src/db"
)

type CommentHandler struct {
	db *db.DB
}

func NewCommentHandler(db *db.DB) *CommentHandler {
	return &CommentHandler{db: db}
}
