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
	"time"

	_ "embed"

	"github.com/gorilla/mux"
)

// TODO: try this out: https://go.dev/blog/slog

//go:embed post-detail.html
var postDetailTemplate string

//go:embed comment/comment-form.html
var commentFormTemplate string

//go:embed comment/comment-list.html
var commentListTemplate string

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

	currentUser, userLoggedIn := auth.ExtractUserFromContext(r.Context())

	var commentListItems []CommentListItem
	for _, commentRecord := range records {
		canEditComment := false
		if userLoggedIn && commentRecord.UserID == currentUser.ID {
			const createdAtFmt = "2006-01-02T15:04:05Z"
			if t, err := time.Parse(createdAtFmt, commentRecord.CreatedAt); err == nil {
				canEditComment = time.Since(t) < 1*time.Minute
			}
		}
		commentListItems = append(commentListItems, CommentListItem{
			ID:        int(commentRecord.ID),
			Text:      commentRecord.Text,
			UserID:    commentRecord.UserID,
			CreatedAt: date.FormatDate(commentRecord.CreatedAt),
			UserName:  commentRecord.UserName,
			CanEdit:   canEditComment,
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

	// Determine if the current user can edit (owner + within 5 minutes of posting)
	if userLoggedIn {
		const createdAtFormat = "2006-01-02T15:04:05Z"
		if t, err := time.Parse(createdAtFormat, record.CreatedAt); err == nil {
			postView.CanEdit = currentUser.ID == record.FUserID && time.Since(t) < 5*time.Minute
		}
		postView.UserID = currentUser.ID
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

	render.RenderTemplate(
		w,
		"post-detail.html",
		pageData,
		r.Context(),
		postDetailTemplate,
		commentFormTemplate,
		commentListTemplate,
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
	CanEdit          bool
	UserID           string
}

type CommentListItem struct {
	ID        int
	Text      string
	UserID    string
	CreatedAt string
	UserName  string
	CanEdit   bool
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

const editWindowDuration = 5 * time.Minute
const createdAtFormat = "2006-01-02T15:04:05Z"

func (ph *PostHandler) PostEditPOSTHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.ExtractUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Re-fetch post to verify ownership and time window
	record, err := ph.QueryOnePost(postID)
	if err != nil || record == (PostDetailRecord{}) {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	if record.FUserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	t, err := time.Parse(createdAtFormat, record.CreatedAt)
	if err != nil || time.Since(t) >= editWindowDuration {
		http.Error(w, "Edit window expired (5 minutes)", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	rec := PostEditRecord{
		Title:       sanitize.Sanitize(r.FormValue("title")),
		Description: sanitize.Sanitize(r.FormValue("description")),
	}

	if rec.Title == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}

	_, err = ph.UpdatePostTitleAndDescription(postID, user.ID, rec)
	if err != nil {
		log.Error.Printf("msg='could not update post' postID='%d' err='%s'\n", postID, err.Error())
		http.Error(w, "Could not update post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/posts/"+vars["id"], http.StatusSeeOther)
}

func (ph *PostHandler) CommentEditPOSTHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.ExtractUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	postID := vars["id"]
	commentID, err := strconv.Atoi(vars["commentID"])
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	record, err := ph.ch.QueryCommentByID(commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if record.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	t, err := time.Parse(createdAtFormat, record.CreatedAt)
	if err != nil || time.Since(t) >= 1*time.Minute {
		http.Error(w, "Edit window expired (1 minute)", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	text := sanitize.Sanitize(r.FormValue("text"))
	if text == "" {
		http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
		return
	}

	_, err = ph.ch.UpdateCommentText(commentID, user.ID, text)
	if err != nil {
		log.Error.Printf("msg='could not update comment' commentID='%d' err='%s'\n", commentID, err.Error())
		http.Error(w, "Could not update comment", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/posts/"+postID+"/#comment-"+strconv.Itoa(commentID), http.StatusSeeOther)
}
