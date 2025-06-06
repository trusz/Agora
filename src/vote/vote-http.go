package vote

import (
	"agora/src/log"
	"agora/src/server/auth"
	"database/sql"
	"net/http"
	"strconv"
)

func (vh *VoteHandler) VotePOSTHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.ExtractUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	postIDStr := r.FormValue("post_id")
	commentIDStr := r.FormValue("comment_id")

	postID := sql.NullInt64{}

	if postIDStr != "" {
		postIDint, err := strconv.Atoi(postIDStr)
		if err != nil {
			log.Error.Printf("msg='could not convert post_id from string to int' postID='%s'\n", postIDStr)
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}
		postID = sql.NullInt64{
			Int64: int64(postIDint),
			Valid: true,
		}
	}

	commentID := sql.NullInt64{}
	if commentIDStr != "" {
		commentIDint, err := strconv.Atoi(commentIDStr)
		if err != nil {
			log.Error.Printf("msg='could not convert comment_id from string to int' commentID='%s'\n", commentIDStr)
			http.Error(w, "Invalid comment ID", http.StatusBadRequest)
			return
		}
		commentID = sql.NullInt64{
			Int64: int64(commentIDint),
			Valid: true,
		}
	}

	nrOfVotes, err := vh.QueryNrOfVotesPerPostAndUser(postID.Int64, user.ID)
	if err != nil {
		log.Error.Printf("msg='could not query number of votes' err='%s'\n", err.Error())
		http.Error(w, "Could not vote", http.StatusInternalServerError)
		return
	}

	if nrOfVotes > 0 {
		log.Error.Printf("msg='user already voted' user='%s' postID='%d' commentID='%d'\n", user.ID, postID.Int64, commentID.Int64)
		http.Error(w, "You have already voted", http.StatusBadRequest)
		return
	}

	newVote := VoteInsertRecord{
		UserID:    user.ID,
		PostID:    postID,
		CommentID: commentID,
	}

	_, err = vh.InsertNewVote(newVote)
	if err != nil {
		log.Error.Printf("msg='could not insert new vote' err='%s'\n", err.Error())
		http.Error(w, "Could not process vote", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/posts#post-"+strconv.FormatInt(postID.Int64, 10), http.StatusSeeOther)
}
