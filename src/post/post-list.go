package post

import (
	"goazuread/src/render"
	"net/http"
)

func PostListHandler(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "src/post/post-list.html", &render.Page{
		Title: "Posts",
		Data:  nil,
	})
}
