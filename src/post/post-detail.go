package post

import (
	"agora/src/log"
	"agora/src/post/comment"
	"agora/src/render"
	"agora/src/server/auth"
	"agora/src/x/date"
	"agora/src/x/sanitize"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// TODO: try this out: https://go.dev/blog/slog

func (ph *PostHandler) PostDetailGETHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	postIDasString := vars["id"]
	postID, err := strconv.Atoi(postIDasString)
	if err != nil {
		log.Error.Printf("msg='could not convert postID from string to int' postID='%s' err='%s'\n", postIDasString, err.Error())
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// TODO: probably should differentiate between problems
	// so we can send 404 if post not found
	record, err := ph.QueryOnePost(postID)
	if err != nil {
		log.Error.Printf("msg='could not query post by ID' postID='%d' err='%s'\n", postID, err.Error())
		http.Error(w, "Could not retrieve post", http.StatusInternalServerError)
		return
	}

	if record == (PostDetailRecord{}) {
		log.Error.Printf("msg='post not found' postID='%d'\n", postID)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	records, err := ph.ch.QueryAllCommentyByPostID(int(record.ID))
	if err != nil {
		log.Error.Printf("msg='could not query comments for post' postID='%d' err='%s'\n", record.ID, err.Error())
		http.Error(w, "Could not retrieve comments", http.StatusInternalServerError)
		return
	}

	var commentListItems []CommentListItem
	for _, commentRecord := range records {
		commentListItems = append(commentListItems, CommentListItem{
			ID:        int(commentRecord.ID),
			Text:      commentRecord.Text,
			UserID:    commentRecord.UserID,
			CreatedAt: date.FormatDate(commentRecord.CreatedAt),
			UserName:  commentRecord.UserName,
		})
	}

	postView := PostDetailItem{
		ID:               int(record.ID),
		Title:            record.Title,
		URL:              record.URL.String,
		Description:      record.Description,
		CreatedAt:        date.FormatDate(record.CreatedAt),
		UserName:         record.FUserName,
		NumberOFComments: record.FNrOfComments,
	}

	pageData := &render.Page{
		Title: "Post: " + record.Title,
		Data: struct {
			Post     PostDetailItem
			Comments []CommentListItem
		}{
			Post:     postView,
			Comments: commentListItems,
		},
	}

	// log.Pretty("pageData", pageData)

	render.RenderTemplate(
		w,
		"src/post/post-detail.html",
		pageData,
		r.Context(),
		"src/post/comment/comment-form.html",
		"src/post/comment/comment-list.html",
	)
}

func (ph *PostHandler) PostDetailDELETEHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.ExtractUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	varPostID := vars["id"]

	postID, err := strconv.Atoi(varPostID)
	if err != nil {
		log.Error.Printf(
			"msg='could not convert postid from string to int' postID='%s'\n",
			varPostID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := ph.deletePost(postID, user.ID); err != nil {
		log.Error.Printf("msg='could not delete post' postID='%d' userID='%s' err='%s'\n", postID, user.ID, err.Error())
		http.Error(w, "Could not delete post", http.StatusInternalServerError)
		return
	}

	ph.ch.RemoveAllCommentsOfPost(postID)
	// TODO: should remove votes,
	// but the vote handler already uses post handler
	// so we cannot create a circular dependency
	// Move interaction to a channel and messages?
	// ph.vh.RemoveAllVotesOfPost(postID)

	http.Redirect(w, r, "/posts/", http.StatusSeeOther)
}

type PostDetailItem struct {
	ID               int
	Title            string
	URL              string
	Description      string
	CreatedAt        string
	UserName         string
	NumberOFComments int
}

type CommentListItem struct {
	ID        int
	Text      string
	UserID    string
	CreatedAt string
	UserName  string
}

func (ph *PostHandler) PostCommentPOSTHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.ExtractUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	varPostID := vars["id"]

	postID, err := strconv.Atoi(varPostID)
	if err != nil {
		log.Error.Printf(
			"msg='could not convert postid from string to int' postID='%s'\n",
			varPostID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	newComment := comment.CommentInsertRecord{
		Text:   sanitize.Sanitize(r.FormValue("comment")),
		PostID: postID,
		UserID: user.ID,
	}

	newCommentID, err := ph.ch.InsertNewComment(newComment)
	if err != nil {
		log.Error.Printf("msg='could not add new comment' postID='%d' err='%s'\n", postID, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := "/posts/" + varPostID + "/#comment-" + strconv.Itoa(int(newCommentID))
	http.Redirect(w, r, url, http.StatusSeeOther)
}
