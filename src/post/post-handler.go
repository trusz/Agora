package post

import (
	"agora/src/db"
	"agora/src/log"
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

func (ph *PostHandler) FindPostByID(postID int) Post {
	postRecord, err := ph.QueryOnePost(postID)
	if err != nil {
		log.Error.Printf("msg='could not query post by ID' postID='%d' err='%s'\n", postID, err.Error())
		return PostNull
	}

	return postRecordToPost(postRecord)
}

func (ph *PostHandler) FindAllPosts() []Post {
	postRecords, err := ph.QueryAllPosts()
	if err != nil {
		log.Error.Printf("msg='could not query all posts' err='%s'\n", err.Error())
		return nil
	}
	posts := make([]Post, 0, len(postRecords))
	for _, postRecord := range postRecords {
		posts = append(posts, postRecordToPost(postRecord))
	}
	return posts
}

func postRecordToPost(pr PostRecord) Post {
	return Post{
		ID:               int(pr.ID),
		Title:            pr.Title,
		URL:              pr.URL.String,
		Description:      pr.Description,
		CreatedAt:        pr.CreatedAt,
		UserID:           pr.FUserID,
		UserName:         pr.FUserName,
		NumberOFComments: pr.FNrOfComments,
	}
}
