package post

import (
	"agora/src/log"
	"agora/src/render"
	"agora/src/server/auth"
	"agora/src/x/date"
	"math"
	"net/http"
	"strconv"

	"golang.org/x/net/context"
)

func (ph *PostHandler) PostListHandler(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	user, ok := auth.ExtractUserFromContext(context)
	if !ok {
		log.Error.Printf("msg='could not get user from context' context='%#v'\n", context)
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	records, err := ph.QueryAllPostsForTheList(user.ID)
	if err != nil {
		log.Error.Printf("msg='could not query all posts' err='%s'\n", err.Error())
		http.Error(w, "Could not retrieve posts", http.StatusInternalServerError)
	}

	if len(records) == 0 {
		ph.renderList(w, nil, 1, 1, r.Context())
		return
	}

	sizeOfPage := 5
	page := 1
	totalPages := int(math.Ceil(float64(len(records)) / float64(sizeOfPage)))

	qPage := r.URL.Query().Get("page")
	if qPage != "" {
		var err error
		page, err = strconv.Atoi(qPage)
		if err != nil {
			log.Error.Printf("msg='could not convert page from string to int' page='%s'\n", qPage)
			http.Error(w, "Invalid page number", http.StatusBadRequest)
			return
		}
		if page < 1 {
			log.Error.Printf("msg='page number is less than 1' page='%d'\n", page)
			http.Error(w, "Page number must be greater than 0", http.StatusBadRequest)
			return
		}
		if page > totalPages {
			log.Error.Printf("msg='page number is greater than total pages' page='%d' totalPages='%d'\n", page, totalPages)
			http.Error(w, "Page number exceeds total pages", http.StatusBadRequest)
			return
		}
	}
	cutFrom := (page - 1) * sizeOfPage
	cutTo := cutFrom + sizeOfPage
	finalCutTo := int(math.Min(float64(len(records)), float64(cutTo)))

	records = records[cutFrom:finalCutTo]

	var postListItems []PostListItem
	for _, record := range records {

		descCutLength := 100

		CutOfDescription := record.Description
		if len(CutOfDescription) > descCutLength {
			CutOfDescription = CutOfDescription[:descCutLength] + " …"
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

	ph.renderList(w, postListItems, page, totalPages, r.Context())
}

func (ph *PostHandler) renderList(
	w http.ResponseWriter,
	postListItems []PostListItem,
	page int,
	totalPages int,
	ctx context.Context,
) {
	render.RenderTemplate(
		w,
		"src/post/post-list.html",
		&render.Page{
			Title: "Posts",
			Data: struct {
				Posts       []PostListItem
				HasPrevPage bool
				HasNextPage bool
				PrevPage    int
				NextPage    int
			}{
				Posts:       postListItems,
				HasPrevPage: page > 1,
				HasNextPage: page < totalPages,
				PrevPage:    page - 1,
				NextPage:    page + 1,
			},
		},
		ctx,
	)
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
