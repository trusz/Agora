package post

import (
	"agora/src/log"
	"database/sql"
)

// type PostRecord struct {
// 	ID          int64
// 	Title       string
// 	URL         sql.NullString
// 	Description string
// 	CreatedAt   string

// 	FUserID       string
// 	FUserName     string
// 	FNrOfComments int
// }

type PostRecord struct {
	ID          int64
	Title       string
	URL         sql.NullString
	Description string
	CreatedAt   string
	Rank        int
}

const TABLE_QUERY = `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT UNIQUE,
		description TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		fk_user_id TEXT NOT NULL,
		rank INT DEFAULT 0,

		CONSTRAINT "fk_user_id" FOREIGN KEY("fk_user_id") REFERENCES users(id)
	);
	`

func (ph *PostHandler) CreateDBTable() error {
	// Create the posts table if it doesn't exist
	_, err := ph.db.Exec(TABLE_QUERY)
	if err != nil {
		log.Error.Printf("Error creating posts table: %v", err)
		return err
	}
	return nil
}

func (ph *PostHandler) InsertNewPost(record PostNewRecord) (int64, error) {
	// Insert a new post into the database
	var url interface{}
	if record.URL == "" {
		url = nil
	} else {
		url = record.URL
	}

	result, err := ph.db.Exec(
		`INSERT 
			INTO posts (title, url, description, fk_user_id) 
			VALUES (?, ?, ?, ?)
		`,
		record.Title,
		url,
		record.Description,
		record.UserID,
	)

	if err != nil {
		log.Error.Printf("Error inserting new post: %v", err)
		return 0, err
	}

	return result.LastInsertId()
}

type PostNewRecord struct {
	Title       string
	URL         string
	Description string
	UserID      string
}

func (ph *PostHandler) QueryOnePost(id int) (PostDetailRecord, error) {
	rows, err := ph.db.Query(
		`
		SELECT 
			p.id, p.title, p.url, p.description, p.created_at, p.rank, p.fk_user_id,
			u.name,
			(SELECT count(*) FROM comments c WHERE fk_post_id=p.id ) nr_comments,
			(SELECT count(*) FROM votes v WHERE fk_post_id=p.id ) nr_votes
		FROM posts p
		LEFT JOIN users u ON u.id = p.fk_user_id
		WHERE p.id = ?`,
		id,
	)
	if err != nil {
		log.Error.Println("Could not query post with id=", id)
		return PostDetailRecord{}, err
	}
	defer rows.Close()

	for rows.Next() {

		var record PostDetailRecord
		err := rows.Scan(
			&record.ID,
			&record.Title,
			&record.URL,
			&record.Description,
			&record.CreatedAt,
			&record.Rank,
			&record.FUserID,
			&record.FUserName,
			&record.FNrOfComments,
			&record.FNrOfVotes,
		)

		if err != nil {
			log.Error.Printf("Could not scan post: id=%d error=%v", id, err)
			return PostDetailRecord{}, err
		}

		return record, nil
	}

	return PostDetailRecord{}, nil

}

type PostDetailRecord struct {
	PostRecord
	FUserID       string
	FUserName     string
	FNrOfComments int
	FNrOfVotes    int
}

func (ph *PostHandler) QueryAllPostsForTheList(userID string) ([]PostListRecord, error) {

	// Query all posts from the database
	rows, err := ph.db.Query(`
		SELECT 
			p.id, p.title, p.url, p.description, p.created_at, p.rank,
			u.name,
			(Select count(*) from comments c where fk_post_id=p.id ) nr_comments,
			(Select count(*) from votes v where fk_post_id=p.id ) nr_votes,
			(select count(*) > 0 from votes v where v.fk_post_id = p.id and v.fk_user_id = ?) user_voted,
			p.fk_user_id = ? is_user_author
		FROM posts p
		LEFT JOIN users u ON u.id = p.fk_user_id
		ORDER BY p.rank DESC, p.created_at DESC
	`,
		userID,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []PostListRecord
	for rows.Next() {

		var record PostListRecord
		err := rows.Scan(
			&record.ID,
			&record.Title,
			&record.URL,
			&record.Description,
			&record.CreatedAt,
			&record.Rank,
			&record.FUserName,
			&record.FNrOfComments,
			&record.FNrOfVotes,
			&record.UserVoted,
			&record.UserIsAuthor,
		)
		if err != nil {
			log.Error.Printf("Could not scan post: error=%v rows=%v", err, rows)
			continue
		}

		records = append(records, record)
	}

	return records, nil
}

func (ph *PostHandler) QueryAllPostsForRanking() ([]PostForRanking, error) {

	// Query all posts from the database
	rows, err := ph.db.Query(`
		SELECT 
			p.id, p.title, p.url, p.description, p.created_at, p.rank,
			(Select count(*) from votes v where fk_post_id=p.id ) nr_votes
		FROM posts p
	`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []PostForRanking
	for rows.Next() {

		var record PostForRanking
		err := rows.Scan(
			&record.ID,
			&record.Title,
			&record.URL,
			&record.Description,
			&record.CreatedAt,
			&record.Rank,
			&record.FNrOfVotes,
		)
		if err != nil {
			log.Error.Printf("Could not scan post: error=%v rows=%v", err, rows)
			continue
		}

		records = append(records, record)
	}

	return records, nil
}

type PostListRecord struct {
	PostRecord
	FUserName     string
	FNrOfComments int
	FNrOfVotes    int
	FNUserVoted   int
	UserVoted     int
	UserIsAuthor  int
}

type PostForRanking struct {
	PostRecord
	FNrOfVotes int
}

func (ph *PostHandler) UpdateRank(postID int64, rank int) error {
	// Update the rank of a post
	_, err := ph.db.Exec(
		`UPDATE posts SET rank = ? WHERE id = ?`,
		rank,
		postID,
	)
	if err != nil {
		log.Error.Printf("Error updating post rank: %v", err)
		return err
	}
	return nil
}

func (ph *PostHandler) deletePost(postID int, userID string) error {
	// Delete a post from the database
	_, err := ph.db.Exec(
		`DELETE FROM posts WHERE id = ? AND fk_user_id = ?`,
		postID,
		userID,
	)
	if err != nil {
		log.Error.Printf("Error deleting post: %v", err)
		return err
	}
	return nil
}
