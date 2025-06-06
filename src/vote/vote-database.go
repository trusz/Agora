package vote

import (
	"agora/src/log"
	"database/sql"
)

type VoteRecord struct {
	PostID    int64
	CommentID int64
	UserID    string
	CreatedAt string
}

const TABLE_QUERY = `CREATE TABLE IF NOT EXISTS votes (
		fk_post_id INTEGER,
		fk_comment_id INTEGER,
		fk_user_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

		UNIQUE(fk_post_id, fk_comment_id, fk_user_id),
		CONSTRAINT "fk_post_id" FOREIGN KEY("fk_post_id") REFERENCES posts(id),
		CONSTRAINT "fk_comment_id" FOREIGN KEY("fk_comment_id") REFERENCES comments(id),
		CONSTRAINT "fk_user_id" FOREIGN KEY("fk_user_id") REFERENCES users(id),
		CHECK (
        (fk_post_id IS NOT NULL AND fk_comment_id IS NULL) OR
        (fk_post_id IS NULL AND fk_comment_id IS NOT NULL)
    	)
	);
	`

func (vh *VoteHandler) QueryNrOfVotesPerPostAndUser(postID int64, userID string) (int, error) {
	var count int64
	err := vh.db.QueryRow(
		`SELECT COUNT(*) FROM votes 
		 WHERE fk_post_id = ? AND fk_user_id = ?`,
		postID,
		userID,
	).Scan(&count)
	if err != nil {
		log.Error.Printf("Error querying number of votes: %v", err)
		return 0, err
	}
	return int(count), nil
}

func (vh *VoteHandler) CreateDBTable() error {
	// Create the votes table if it doesn't exist
	_, err := vh.db.Exec(TABLE_QUERY)
	if err != nil {
		log.Error.Printf("Error creating votes table: %v", err)
		return err
	}
	return nil
}

func (vh *VoteHandler) InsertNewVote(record VoteInsertRecord) (int64, error) {
	// Insert a new vote into the database
	result, err := vh.db.Exec(
		`INSERT INTO votes (fk_post_id, fk_comment_id, fk_user_id) 
		 VALUES (?, ?, ?)`,
		record.PostID,
		record.CommentID,
		record.UserID,
	)
	if err != nil {
		log.Error.Printf("Error inserting new vote: %v", err)
		return 0, err
	}
	return result.LastInsertId()
}

type VoteInsertRecord struct {
	PostID    sql.NullInt64
	CommentID sql.NullInt64
	UserID    string
}
