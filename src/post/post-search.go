package post

import (
	"agora/src/log"
	"agora/src/render"
	"agora/src/server/auth"
	"agora/src/x/date"
	_ "embed"
	"net/http"
	"strconv"

	"golang.org/x/net/context"
)

// SearchResultItem extends PostListItem with search-specific metadata
type SearchResultItem struct {
	PostListItem
	MatchType    string // "title", "description", or "comment"
	MatchSnippet string // Highlighted excerpt of the match
}

func (ph *PostHandler) PostSearchGETHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := auth.ExtractUserFromContext(ctx)
	if !ok {
		log.Error.Printf("msg='could not get user from context' context='%#v'\n", ctx)
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	pageStr := r.URL.Query().Get("page")

	// Parse page number (default to 1)
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Pagination settings
	const resultsPerPage = 20
	offset := (page - 1) * resultsPerPage

	var searchResults []SearchResultItem
	var totalResults int

	if query != "" && len(query) >= 3 {
		// Try FTS5 search first, fall back to LIKE-based search if FTS5 is not available
		results, err := ph.SearchPostsFTS5(query, user.ID, resultsPerPage, offset)
		if err != nil {
			// FTS5 not available or error, use fallback search
			log.Info.Printf("msg='FTS5 search failed, using fallback' query='%s' err='%s'\n", query, err.Error())
			results, err = ph.SearchPostsFallback(query, user.ID, resultsPerPage, offset)
			if err != nil {
				log.Error.Printf("msg='could not search posts' query='%s' err='%s'\n", query, err.Error())
				http.Error(w, "Search failed", http.StatusInternalServerError)
				return
			}

			// Get total count for pagination (fallback)
			totalResults, err = ph.CountSearchResultsFallback(query)
			if err != nil {
				log.Error.Printf("msg='could not count search results' query='%s' err='%s'\n", query, err.Error())
			}
		} else {
			// FTS5 search succeeded
			// Get total count for pagination
			totalResults, err = ph.CountSearchResults(query)
			if err != nil {
				log.Error.Printf("msg='could not count search results' query='%s' err='%s'\n", query, err.Error())
			}
		}

		for _, result := range results {
			item := SearchResultItem{
				PostListItem: PostListItem{
					ID:               int(result.ID),
					Title:            result.Title,
					URL:              result.URL.String,
					Description:      result.Description,
					CreatedAt:        date.FormatDate(result.CreatedAt),
					UserName:         result.FUserName,
					NumberOfComments: result.FNrOfComments,
					NumberOfVotes:    result.FNrOfVotes,
					UserVoted:        result.UserVoted == 1,
					UserIsAuthor:     result.UserIsAuthor == 1,
				},
				MatchType:    result.MatchType,
				MatchSnippet: result.MatchSnippet,
			}
			searchResults = append(searchResults, item)
		}
	}

	// Calculate pagination info
	totalPages := (totalResults + resultsPerPage - 1) / resultsPerPage
	if totalPages < 1 {
		totalPages = 1
	}

	ph.renderSearch(w, searchResults, query, page, totalPages, totalResults, ctx)
}

//go:embed post-search.html
var postSearchTemplate string

func (ph *PostHandler) renderSearch(
	w http.ResponseWriter,
	results []SearchResultItem,
	query string,
	currentPage int,
	totalPages int,
	totalResults int,
	ctx context.Context,
) {
	render.RenderTemplate(
		w,
		"post-search.html",
		&render.Page{
			Title: "Search",
			Data: struct {
				Results      []SearchResultItem
				Query        string
				CurrentPage  int
				TotalPages   int
				TotalResults int
			}{
				Results:      results,
				Query:        query,
				CurrentPage:  currentPage,
				TotalPages:   totalPages,
				TotalResults: totalResults,
			},
		},
		ctx,
		postSearchTemplate,
	)
}
