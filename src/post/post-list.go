package post

import (
	"agora/src/log"
	"agora/src/render"
	"agora/src/x/date"
	"net/http"
)

func (ph *PostHandler) PostListHandler(w http.ResponseWriter, r *http.Request) {

	records, err := ph.QueryAllPosts()
	if err != nil {
		log.Error.Printf("msg='could not query all posts' err='%s'\n", err.Error())
		http.Error(w, "Could not retrieve posts", http.StatusInternalServerError)
	}
	var postListItesms []PostListItem
	for _, record := range records {

		CutOfDescription := record.Description
		if len(CutOfDescription) > 100 {
			CutOfDescription = CutOfDescription[:100] + " …"
		}

		postListItesms = append(postListItesms, PostListItem{
			ID:               int(record.ID),
			Title:            record.Title,
			URL:              record.URL.String,
			Description:      CutOfDescription,
			CreatedAt:        date.FormatDate(record.CreatedAt),
			UserName:         record.FUserName,
			NumberOfComments: record.FNrOfComments,
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
