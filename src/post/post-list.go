package post

import (
	"goazuread/src/render"
	"net/http"
)

func (ph *PostHandler) PostListHandler(w http.ResponseWriter, r *http.Request) {

	posts, _ := ph.QueryAllPosts()

	for i := range posts {
		if len(posts[i].Description) > 100 {
			posts[i].Description = posts[i].Description[:100] + " …"
		}
	}

	render.RenderTemplate(w, "src/post/post-list.html", &render.Page{
		Title: "Posts",
		Data:  posts,
	})
}
