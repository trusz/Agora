package post

import (
	"agora/src/log"
	"agora/src/render"
	"agora/src/x/sanitize"
	"net/http"
	"strconv"
)

func (ph *PostHandler) PostSubmitGETHandler(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "src/post/post-submit.html", &render.Page{
		Title: "Submit Post",
		Data:  nil,
	})
}

func (ph *PostHandler) PostSubmitPOSTHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := sanitize.Sanitize(r.FormValue("title"))
	url := sanitize.Sanitize(r.FormValue("url"))
	desc := sanitize.Sanitize(r.FormValue("description"))

	newPost := Post{
		Title:       title,
		URL:         url,
		Description: desc,
	}

	newPostID, err := ph.InsertNewPost(newPost)
	if err != nil {
		log.Error.Printf("msg='could not create new post' err='%s'\n", err.Error())
		http.Error(w, "Could not insert create post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/posts#post-"+strconv.Itoa(int(newPostID)), http.StatusSeeOther)
}
