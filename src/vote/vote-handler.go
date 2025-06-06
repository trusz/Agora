package vote

import "agora/src/db"

type VoteHandler struct {
	db *db.DB
}

func NewVoteHandler(db *db.DB) *VoteHandler {
	return &VoteHandler{
		db: db,
	}
}
