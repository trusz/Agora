package vote

import (
	"agora/src/db"
	"agora/src/post"
)

type VoteHandler struct {
	db *db.DB
	ph *post.PostHandler
}

func NewVoteHandler(db *db.DB, ph *post.PostHandler) *VoteHandler {
	return &VoteHandler{
		db: db,
		ph: ph,
	}
}
