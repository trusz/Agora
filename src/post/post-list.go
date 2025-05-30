package post

import (
	"agora/src/log"
	"agora/src/render"
	"net/http"
)

func (ph *PostHandler) PostListHandler(w http.ResponseWriter, r *http.Request) {

	posts, err := ph.QueryAllPosts()
	if err != nil {
		log.Error.Printf("msg='could not query all posts' err='%s'\n", err.Error())
		http.Error(w, "Could not query posts: "+err.Error(), http.StatusInternalServerError)
		return
	}

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
