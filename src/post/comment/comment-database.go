package comment

import "agora/src/log"

const TABLE_QUERY = `
	CREATE TABLE IF NOT EXISTS "comments" (
		"id"	INTEGER,
		"text"	TEXT NOT NULL,
		"created_at" DATETIME DEFAULT CURRENT_TIMESTAMP,
		"fk_post_id" INTEGER,
		"fk_user_id" INTEGER,

		CONSTRAINT "fk_post id" FOREIGN KEY("fk_post_id") REFERENCES posts(id),
		PRIMARY KEY("id" AUTOINCREMENT)
	);
`

func (ch *CommentHandler) CreateDBTable() error {
	_, err := ch.db.Exec(TABLE_QUERY)
	if err != nil {
		log.Error.Printf("Error creating posts table: %v", err)
		return err
	}
	return nil
}

func (ch *CommentHandler) InsertNewComment(c Comment) (int64, error) {
	// Insert a new post into the database
	result, err := ch.db.Exec("INSERT INTO comments (text, fk_post_id, fk_user_id) VALUES (?, ?, ?)", c.Text, c.PostID, -1)
	if err != nil {
		log.Error.Printf("Error inserting new comment: %v", err)
		return 0, err
	}
	return result.LastInsertId()
}

func (ch *CommentHandler) QueryAllCommentyByPostID(postID int) ([]Comment, error) {
	rows, err := ch.db.Query("SELECT id, text, fk_user_id, created_at from comments where fk_post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var id int64
		var text string
		var userID int
		var createdAt string

		if err := rows.Scan(&id, &text, &userID, &createdAt); err != nil {
			log.Error.Printf("msg='could not scan row' err='%s'\n", err)
			return nil, err
		}

		comment := Comment{
			ID:        int(id),
			Text:      text,
			PostID:    postID,
			UserID:    userID,
			CreatedAt: createdAt,
		}
		comments = append(comments, comment)
	}

	return comments, nil

}
