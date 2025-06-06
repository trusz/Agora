package post

import (
	"agora/src/log"
	"agora/src/render"
	"agora/src/server/auth"
	"agora/src/x/date"
	"net/http"
)

func (ph *PostHandler) PostListHandler(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	user, ok := auth.ExtractUserFromContext(context)
	if !ok {
		log.Error.Printf("msg='could not get user from context' context='%#v'\n", context)
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	records, err := ph.QueryAllPosts(user.ID)
	if err != nil {
		log.Error.Printf("msg='could not query all posts' err='%s'\n", err.Error())
		http.Error(w, "Could not retrieve posts", http.StatusInternalServerError)
	}

	var postListItems []PostListItem
	for _, record := range records {

		cutLength := 100

		CutOfDescription := record.Description
		if len(CutOfDescription) > cutLength {
			CutOfDescription = CutOfDescription[:cutLength] + " …"
		}

		postListItems = append(postListItems, PostListItem{
			ID:               int(record.ID),
			Title:            record.Title,
			URL:              record.URL.String,
			Description:      CutOfDescription,
			CreatedAt:        date.FormatDate(record.CreatedAt),
			UserName:         record.FUserName,
			NumberOfComments: record.FNrOfComments,
			NumberOfVotes:    record.FNrOfVotes,
			UserVoted:        record.UserVoted == 1,
		})

	}

	render.RenderTemplate(w, "src/post/post-list.html", &render.Page{
		Title: "Posts",
		Data:  postListItems,
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
	NumberOfVotes    int
	UserVoted        bool
}
