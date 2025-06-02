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

	post := ph.FindPostByID(postID)
	if post == PostNull {
		log.Error.Printf("msg='post not found' postID='%d'\n", postID)
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
		return
	}

	comments, err := ph.ch.QueryAllCommentyByPostID(post.ID)
	if err != nil {
		log.Error.Printf("msg='could not query comments for post' postID='%d' err='%s'\n", post.ID, err.Error())
		http.Error(w, "Could not retrieve comments", http.StatusInternalServerError)
		return
	}

	postView := PostView{
		ID:               post.ID,
		Title:            post.Title,
		URL:              post.URL,
		Description:      post.Description,
		CreatedAt:        date.FormatDate(post.CreatedAt),
		UserName:         post.UserName,
		NumberOFComments: post.NumberOFComments,
	}

	pageData := &render.Page{
		Title: "Post: " + post.Title,
		Data: struct {
			Post     PostView
			Comments []comment.Comment
		}{
			Post:     postView,
			Comments: comments,
		},
	}

	// log.Pretty("pageData", pageData)

	render.RenderTemplate(
		w,
		"src/post/post-detail.html",
		pageData,
		"src/post/comment/comment-form.html",
		"src/post/comment/comment-list.html",
	)
}

type PostView struct {
	ID               int
	Title            string
	URL              string
	Description      string
	CreatedAt        string
	UserName         string
	NumberOFComments int
}

func (ph *PostHandler) PostCommentPOSTHandler(w http.ResponseWriter, r *http.Request) {
	loggedInUser, ok := auth.ExtractUserFromContext(r.Context())
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

	newCommentText := sanitize.Sanitize(r.FormValue("comment"))

	newCommentID, err := ph.ch.AddNewComment(postID, newCommentText, loggedInUser.ID)
	if err != nil {
		log.Error.Printf("msg='could not add new comment' postID='%d' err='%s'\n", postID, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/posts/"+varPostID+"/#comment-"+strconv.Itoa(newCommentID), http.StatusSeeOther)
}
