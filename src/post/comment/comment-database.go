package comment

import (
	"agora/src/log"
	"agora/src/x/sanitize"
)

const TABLE_QUERY = `
	CREATE TABLE IF NOT EXISTS "comments" (
		"id"	INTEGER,
		"text"	TEXT NOT NULL,
		"created_at" DATETIME DEFAULT CURRENT_TIMESTAMP,
		"fk_post_id" INTEGER,
		"fk_user_id" TEXT NOT NULL,

		CONSTRAINT "fk_post_id" FOREIGN KEY("fk_post_id") REFERENCES posts(id),
		CONSTRAINT "fk_user_id" FOREIGN KEY("fk_user_id") REFERENCES users(id),
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
	result, err := ch.db.Exec("INSERT INTO comments (text, fk_post_id, fk_user_id) VALUES (?, ?, ?)", c.Text, c.PostID, c.UserID)
	if err != nil {
		log.Error.Printf("Error inserting new comment: %v", err)
		return 0, err
	}
	return result.LastInsertId()
}

func (ch *CommentHandler) QueryAllCommentyByPostID(postID int) ([]Comment, error) {
	rows, err := ch.db.Query(
		`SELECT c.id, c.text, c.fk_user_id, c.created_at, u.name
		 FROM comments c
		 LEFT JOIN users u ON u.id = c.fk_user_id
		 WHERE c.fk_post_id = ?`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var id int64
		var text string
		var userID string
		var createdAt string
		var userName string

		if err := rows.Scan(&id, &text, &userID, &createdAt, &userName); err != nil {
			log.Error.Printf("msg='could not scan row' err='%s'\n", err)
			return nil, err
		}

		comment := Comment{
			ID:        int(id),
			Text:      sanitize.Sanitize(text),
			PostID:    postID,
			UserID:    userID,
			CreatedAt: createdAt,
			UserName:  userName,
		}
		comments = append(comments, comment)
	}

	return comments, nil

}
