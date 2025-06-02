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

func (ph *PostHandler) InsertNewPost(p Post) (int64, error) {
	// Insert a new post into the database
	var url interface{}
	if p.URL == "" {
		url = nil
	} else {
		url = p.URL
	}

	result, err := ph.db.Exec("INSERT INTO posts (title, url, description, fk_user_id) VALUES (?, ?, ?, ?)", p.Title, url, p.Description, p.UserID)
	if err != nil {
		log.Error.Printf("Error inserting new post: %v", err)
		return 0, err
	}
	return result.LastInsertId()
}

func (ph *PostHandler) QueryOnePost(id int) (PostRecord, error) {
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
		return PostRecord{}, err
	}
	defer rows.Close()

	// var wantedPost Post
	for rows.Next() {
		// var id int64
		// var title string
		// var description string
		// var createdAt string
		// var url sql.NullString
		// var userName string
		// var nrComments int
		// Scan the row into variables

		wantedPostRecord, err := scanPostRecord(rows)
		if err != nil {
			slog.Error("Could not scan post", "id", id, "error", err)
			return PostRecord{}, err
		}

		// if err := rows.Scan(&id, &title, &url, &description, &createdAt, &userName, &nrComments); err != nil {
		// 	log.Error.Println("Could not scan post with id=", id)
		// 	return PostRecord{}, err
		// }
		// var urlStr string
		// if url.Valid {
		// 	urlStr = url.String
		// }

		// wantedPost = Post{
		// 	ID:               int(id),
		// 	Title:            sanitize.Sanitize(title),
		// 	URL:              sanitize.Sanitize(urlStr),
		// 	Description:      sanitize.Sanitize(description),
		// 	CreatedAt:        createdAt,
		// 	UserName:         userName,
		// 	NumberOFComments: nrComments,
		// }
		return wantedPostRecord, nil
	}

	return PostRecord{}, nil

}

func (ph *PostHandler) QueryAllPosts() ([]PostRecord, error) {
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

	var postRecords []PostRecord
	for rows.Next() {

		wantedPostRecord, err := scanPostRecord(rows)
		if err != nil {
			slog.Error("Could not scan post", "error", err, "rows", rows)
			continue
		}

		postRecords = append(postRecords, wantedPostRecord)
	}

	return postRecords, nil
}

func scanPostRecord(rows *sql.Rows) (PostRecord, error) {

	var p PostRecord
	err := rows.Scan(
		&p.ID,
		&p.Title,
		&p.URL,
		&p.Description,
		&p.CreatedAt,
		&p.FUserID,
		&p.FUserName,
		&p.FNrOfComments,
	)
	if err != nil {
		return PostRecord{}, err
	}

	return p, nil
}
