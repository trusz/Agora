package post

import (
	"goazuread/src/render"
	"net/http"
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

	title := r.FormValue("title")
	url := r.FormValue("url")
	desc := r.FormValue("description")

	newPost := Post{
		Title:       title,
		URL:         url,
		Description: desc,
	}

	ph.InsertNewPost(newPost)

	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}
