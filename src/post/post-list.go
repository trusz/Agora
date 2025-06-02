package post

import (
	"agora/src/render"
	"agora/src/x/date"
	"net/http"
)

func (ph *PostHandler) PostListHandler(w http.ResponseWriter, r *http.Request) {

	posts := ph.FindAllPosts()
	var postListItesms []PostListItem
	for _, post := range posts {

		CutOfDescription := post.Description
		if len(CutOfDescription) > 100 {
			CutOfDescription = CutOfDescription[:100] + " …"
		}

		postListItesms = append(postListItesms, PostListItem{
			ID:               post.ID,
			Title:            post.Title,
			URL:              post.URL,
			Description:      CutOfDescription,
			CreatedAt:        date.FormatDate(post.CreatedAt),
			UserName:         post.UserName,
			NumberOfComments: post.NumberOFComments,
		})

	}

	render.RenderTemplate(w, "src/post/post-list.html", &render.Page{
		Title: "Posts",
		Data:  postListItesms,
	})
}

type PostListItem struct {
	ID               int
	Title            string
	URL              string
	Description      string
	CreatedAt        string
	UserName         string
	NumberOfComments int
}
