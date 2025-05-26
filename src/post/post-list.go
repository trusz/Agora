package post

import (
	"goazuread/src/render"
	"net/http"
)

func (ph *PostHandler) PostListHandler(w http.ResponseWriter, r *http.Request) {

	posts, _ := ph.QueryAllPosts()

	render.RenderTemplate(w, "src/post/post-list.html", &render.Page{
		Title: "Posts",
		Data:  posts,
	})
}
