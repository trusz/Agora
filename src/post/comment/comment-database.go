package comment

import (
	"agora/src/log"
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

func (ch *CommentHandler) InsertNewComment(c CommentInsertRecord) (int64, error) {
	// Insert a new post into the database
	result, err := ch.db.Exec(
		`INSERT INTO comments (text, fk_post_id, fk_user_id) VALUES (?, ?, ?)`,
		c.Text,
		c.PostID,
		c.UserID,
	)
	if err != nil {
		log.Error.Printf("Error inserting new comment: %v", err)
		return 0, err
	}
	return result.LastInsertId()
}

type CommentInsertRecord struct {
	Text   string
	PostID int
	UserID string
}

func (ch *CommentHandler) QueryAllCommentyByPostID(postID int) ([]CommentListRecord, error) {
	rows, err := ch.db.Query(
		`SELECT c.id, c.text, c.created_at, c.fk_user_id, u.name
		 FROM comments c
		 LEFT JOIN users u ON u.id = c.fk_user_id
		 WHERE c.fk_post_id = ?`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []CommentListRecord
	for rows.Next() {
		var record CommentListRecord

		err := rows.Scan(
			&record.ID,
			&record.Text,
			&record.CreatedAt,
			&record.UserID,
			&record.UserName,
		)
		if err != nil {
			log.Error.Printf("msg='could not scan row' err='%s'\n", err)
			return nil, err
		}

		records = append(records, record)
	}

	return records, nil
}

func (ch *CommentHandler) UpdateCommentText(commentID int, userID string, text string) (bool, error) {
	result, err := ch.db.Exec(
		`UPDATE comments SET text = ? WHERE id = ? AND fk_user_id = ?`,
		text, commentID, userID,
	)
	if err != nil {
		log.Error.Printf("Error updating comment: %v", err)
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (ch *CommentHandler) QueryCommentByID(commentID int) (CommentListRecord, error) {
	var record CommentListRecord
	err := ch.db.QueryRow(
		`SELECT c.id, c.text, c.created_at, c.fk_user_id, u.name
		 FROM comments c
		 LEFT JOIN users u ON u.id = c.fk_user_id
		 WHERE c.id = ?`,
		commentID,
	).Scan(&record.ID, &record.Text, &record.CreatedAt, &record.UserID, &record.UserName)
	if err != nil {
		return CommentListRecord{}, err
	}
	return record, nil
}

type CommentListRecord struct {
	ID        int
	Text      string
	PostID    int
	UserID    string
	CreatedAt string
	UserName  string
}
