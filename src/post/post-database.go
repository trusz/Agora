package post

import (
	"agora/src/log"
	"database/sql"
	"log/slog"
)

type PostRecord struct {
	ID          int64
	Title       string
	URL         sql.NullString
	Description string
	CreatedAt   string

	FUserID       string
	FUserName     string
	FNrOfComments int
}

const TABLE_QUERY = `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT UNIQUE,
		description TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		fk_user_id TEXT NOT NULL,

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
			p.id, p.title, p.url, p.description, p.created_at, p.fk_user_id,
			u.name,
			(Select count(*) from comments c where fk_post_id=p.id ) nr_comments
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
			&record.FUserID,
			&record.FUserName,
			&record.FNrOfComments,
		)

		if err != nil {
			slog.Error("Could not scan post", "id", id, "error", err)
			return PostDetailRecord{}, err
		}

		return record, nil
	}

	return PostDetailRecord{}, nil

}

type PostDetailRecord struct {
	ID            int64
	Title         string
	URL           sql.NullString
	Description   string
	CreatedAt     string
	FUserID       string
	FUserName     string
	FNrOfComments int
}

func (ph *PostHandler) QueryAllPosts() ([]PostListeRecord, error) {
	// Query all posts from the database
	rows, err := ph.db.Query(`
		SELECT 
			p.id, p.title, p.url, p.description, p.created_at, p.fk_user_id,
			u.name,
			(Select count(*) from comments c where fk_post_id=p.id ) nr_comments
		FROM posts p
		LEFT JOIN users u ON u.id = p.fk_user_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []PostListeRecord
	for rows.Next() {

		var record PostListeRecord
		err := rows.Scan(
			&record.ID,
			&record.Title,
			&record.URL,
			&record.Description,
			&record.CreatedAt,
			&record.FUserID,
			&record.FUserName,
			&record.FNrOfComments,
		)
		if err != nil {
			slog.Error("Could not scan post", "error", err, "rows", rows)
			continue
		}

		records = append(records, record)
	}

	return records, nil
}

type PostListeRecord struct {
	ID            int64
	Title         string
	URL           sql.NullString
	Description   string
	CreatedAt     string
	FUserID       string
	FUserName     string
	FNrOfComments int
}
