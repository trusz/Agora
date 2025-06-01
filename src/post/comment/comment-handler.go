package comment

import (
	"agora/src/db"
	"agora/src/log"
)

type CommentHandler struct {
	db *db.DB
}

func NewCommentHandler(db *db.DB) *CommentHandler {
	return &CommentHandler{db: db}

}

func (ch *CommentHandler) AddNewComment(postID int, text string, userID string) (int, error) {
	newComment := Comment{
		Text:   text,
		PostID: postID,
		UserID: userID,
	}

	commentID, err := ch.InsertNewComment(newComment)

	if err != nil {
		log.Error.Printf("msg='could not insert new comment' comment='%#v' err='%s'\n", newComment, err)
		return -1, err

	}

	return int(commentID), nil
}
