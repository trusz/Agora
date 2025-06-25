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

func (vh *VoteHandler) RemoveAllVotesOfPost(postID int64) error {
	// Remove all votes for a specific post
	_, err := vh.db.Exec(
		`DELETE FROM votes WHERE fk_post_id = ?`,
		postID,
	)
	if err != nil {
		return err
	}

	// Update the rank of the post to 0 after removing all votes
	if err := vh.ph.UpdateRank(postID, 0); err != nil {
		return err
	}

	return nil
}
