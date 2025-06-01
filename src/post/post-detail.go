package post

import (
	"agora/src/log"
	"agora/src/post/comment"
	"agora/src/render"
	"agora/src/server/auth"
	"agora/src/x/sanitize"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (ph *PostHandler) PostDetailGETHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	postIDasString := vars["id"]
	log.Debug.Println("getting post with id=", postIDasString)
	post, err := ph.QueryOnePost(postIDasString)
	if err != nil {
		log.Error.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if post == PostNull {
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
		return
	}

	postID, err := strconv.Atoi(postIDasString)
	var comments []comment.Comment
	if err == nil {
		comments, err = ph.ch.QueryAllCommentyByPostID(postID)
	} else {
		log.Error.Printf("msg='could not convert postID from string to int' err='%s'\n", err.Error())
	}

	pageData := &render.Page{
		Title: "Post: " + post.Title,
		Data: struct {
			Post     Post
			Comments []comment.Comment
		}{
			Post:     post,
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
