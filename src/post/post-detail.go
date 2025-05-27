package post

import (
	"goazuread/src/log"
	"goazuread/src/render"
	"net/http"

	"github.com/gorilla/mux"
)

func (ph *PostHandler) PostDetailGETHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	postID := vars["id"]
	log.Debug.Println("getting post with id=", postID)
	post, err := ph.QueryOnePost(postID)
	if err != nil {
		log.Error.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Debug.Printf("found post: %#v\n", post)

	render.RenderTemplate(w, "src/post/post-detail.html", &render.Page{
		Title: "Post Detail",
		Data:  post,
	})
}
