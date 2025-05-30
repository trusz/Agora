package post

import (
	"agora/src/log"
	"agora/src/x/sanitize"
	"database/sql"
)

const TABLE_QUERY = `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT UNIQUE,
		description TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		user_id TEXT NOT NULL
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

	result, err := ph.db.Exec("INSERT INTO posts (title, url, description, owner_id) VALUES (?, ?, ?, ?)", p.Title, url, p.Description, p.OwnerID)
	if err != nil {
		log.Error.Printf("Error inserting new post: %v", err)
		return 0, err
	}
	return result.LastInsertId()
}

func (ph *PostHandler) QueryOnePost(id string) (Post, error) {
	rows, err := ph.db.Query("SELECT id, title, url, description, created_at FROM posts WHERE id = ?", id)
	if err != nil {
		log.Error.Println("Could not query post with id=", id)
		return Post{}, err
	}
	defer rows.Close()

	var wantedPost Post
	for rows.Next() {
		var id int64
		var title string
		var description string
		var createdAt string
		var url sql.NullString

		if err := rows.Scan(&id, &title, &url, &description, &createdAt); err != nil {
			log.Error.Println("Could not scan post with id=", id)
			return PostNull, err
		}
		var urlStr string
		if url.Valid {
			urlStr = url.String
		}

		wantedPost = Post{
			ID:          int(id),
			Title:       sanitize.Sanitize(title),
			URL:         sanitize.Sanitize(urlStr),
			Description: sanitize.Sanitize(description),
			CreatedAt:   createdAt,
		}
		return wantedPost, nil
	}

	log.Debug.Println("no post found")
	return PostNull, nil

}

func (ph *PostHandler) QueryAllPosts() ([]Post, error) {
	// Query all posts from the database
	rows, err := ph.db.Query("SELECT id, title, url, description, created_at FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var id int64
		var title string
		var description string
		var createdAt string
		var url sql.NullString

		if err := rows.Scan(&id, &title, &url, &description, &createdAt); err != nil {
			return nil, err
		}

		var urlStr string
		if url.Valid {
			urlStr = url.String
		}

		log.Debug.Printf("msg='post found' id=%d title='%s' url='%s' description='%s' created_at='%s'\n", id, title, urlStr, description, createdAt)

		post := Post{
			ID:          int(id),
			Title:       sanitize.Sanitize(title),
			URL:         sanitize.Sanitize(urlStr),
			Description: sanitize.Sanitize(description),
			CreatedAt:   createdAt,
		}
		posts = append(posts, post)
	}

	return posts, nil
}
