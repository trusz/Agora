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

func (ch *CommentHandler) RemoveAllCommentsOfPost(postID int) error {
	// Remove all comments for a specific post
	_, err := ch.db.Exec(
		`DELETE FROM comments WHERE fk_post_id = ?`,
		postID,
	)
	if err != nil {
		return err
	}
	return nil
}
