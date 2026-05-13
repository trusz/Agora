package post

import (
	"agora/src/log"
	"database/sql"
	"fmt"
	"strings"
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

type PostEditRecord struct {
	Title       string
	Description string
}

// UpdatePostTitleAndDescription updates title and description for a post owned by userID.
// Returns (true, nil) on success, (false, nil) if no row matched (wrong owner/id).
func (ph *PostHandler) UpdatePostTitleAndDescription(postID int, userID string, rec PostEditRecord) (bool, error) {
	result, err := ph.db.Exec(
		`UPDATE posts SET title = ?, description = ? WHERE id = ? AND fk_user_id = ?`,
		rec.Title, rec.Description, postID, userID,
	)
	if err != nil {
		log.Error.Printf("Error updating post: %v", err)
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
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

func (ph *PostHandler) SearchPosts(query string, userID string) ([]PostListRecord, error) {
	like := "%" + query + "%"
	rows, err := ph.db.Query(`
		SELECT DISTINCT
			p.id, p.title, p.url, p.description, p.created_at, p.rank,
			u.name,
			(SELECT count(*) FROM comments c WHERE c.fk_post_id = p.id) nr_comments,
			(SELECT count(*) FROM votes v WHERE v.fk_post_id = p.id) nr_votes,
			(SELECT count(*) > 0 FROM votes v WHERE v.fk_post_id = p.id AND v.fk_user_id = ?) user_voted,
			p.fk_user_id = ? is_user_author
		FROM posts p
		LEFT JOIN users u ON u.id = p.fk_user_id
		LEFT JOIN comments co ON co.fk_post_id = p.id
		WHERE p.title LIKE ? OR p.description LIKE ? OR co.text LIKE ?
		ORDER BY p.rank DESC, p.created_at DESC
	`,
		userID, userID, like, like, like,
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
			log.Error.Printf("Could not scan search result: error=%v", err)
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// SearchResult represents a single search result with metadata
type SearchResult struct {
	PostListRecord
	MatchType    string  // "title", "description", or "comment"
	MatchSnippet string  // Highlighted excerpt of the match
	Score        float64 // Relevancy score
}

// CreateFTS5Tables creates the FTS5 virtual tables for full-text search
func (ph *PostHandler) CreateFTS5Tables() error {
	// Create FTS5 virtual table for posts
	_, err := ph.db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS posts_fts USING fts5(
			title, 
			description, 
			content='posts', 
			content_rowid='id',
			tokenize='porter unicode61'
		);
	`)
	if err != nil {
		log.Error.Printf("Error creating posts_fts table: %v", err)
		return err
	}

	// Create FTS5 virtual table for comments
	_, err = ph.db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS comments_fts USING fts5(
			text,
			content='comments',
			content_rowid='id',
			tokenize='porter unicode61'
		);
	`)
	if err != nil {
		log.Error.Printf("Error creating comments_fts table: %v", err)
		return err
	}

	// Create triggers to keep FTS tables in sync with posts
	_, err = ph.db.Exec(`
		CREATE TRIGGER IF NOT EXISTS posts_ai AFTER INSERT ON posts BEGIN
			INSERT INTO posts_fts(rowid, title, description) 
			VALUES (new.id, new.title, new.description);
		END;
	`)
	if err != nil {
		log.Error.Printf("Error creating posts_ai trigger: %v", err)
		return err
	}

	_, err = ph.db.Exec(`
		CREATE TRIGGER IF NOT EXISTS posts_ad AFTER DELETE ON posts BEGIN
			DELETE FROM posts_fts WHERE rowid = old.id;
		END;
	`)
	if err != nil {
		log.Error.Printf("Error creating posts_ad trigger: %v", err)
		return err
	}

	_, err = ph.db.Exec(`
		CREATE TRIGGER IF NOT EXISTS posts_au AFTER UPDATE ON posts BEGIN
			DELETE FROM posts_fts WHERE rowid = old.id;
			INSERT INTO posts_fts(rowid, title, description) 
			VALUES (new.id, new.title, new.description);
		END;
	`)
	if err != nil {
		log.Error.Printf("Error creating posts_au trigger: %v", err)
		return err
	}

	// Create triggers to keep FTS tables in sync with comments
	_, err = ph.db.Exec(`
		CREATE TRIGGER IF NOT EXISTS comments_ai AFTER INSERT ON comments BEGIN
			INSERT INTO comments_fts(rowid, text) VALUES (new.id, new.text);
		END;
	`)
	if err != nil {
		log.Error.Printf("Error creating comments_ai trigger: %v", err)
		return err
	}

	_, err = ph.db.Exec(`
		CREATE TRIGGER IF NOT EXISTS comments_ad AFTER DELETE ON comments BEGIN
			DELETE FROM comments_fts WHERE rowid = old.id;
		END;
	`)
	if err != nil {
		log.Error.Printf("Error creating comments_ad trigger: %v", err)
		return err
	}

	_, err = ph.db.Exec(`
		CREATE TRIGGER IF NOT EXISTS comments_au AFTER UPDATE ON comments BEGIN
			DELETE FROM comments_fts WHERE rowid = old.id;
			INSERT INTO comments_fts(rowid, text) VALUES (new.id, new.text);
		END;
	`)
	if err != nil {
		log.Error.Printf("Error creating comments_au trigger: %v", err)
		return err
	}

	log.Info.Println("FTS5 tables and triggers created successfully")
	return nil
}

// RebuildFTS5Index rebuilds the FTS5 index from existing data
func (ph *PostHandler) RebuildFTS5Index() error {
	// Clear existing FTS data
	_, _ = ph.db.Exec(`DELETE FROM posts_fts;`)
	_, _ = ph.db.Exec(`DELETE FROM comments_fts;`)

	// Populate posts FTS index
	_, err := ph.db.Exec(`
		INSERT INTO posts_fts(rowid, title, description)
		SELECT id, title, description FROM posts;
	`)
	if err != nil {
		log.Error.Printf("Error rebuilding posts FTS index: %v", err)
		return err
	}

	// Populate comments FTS index
	_, err = ph.db.Exec(`
		INSERT INTO comments_fts(rowid, text)
		SELECT id, text FROM comments;
	`)
	if err != nil {
		log.Error.Printf("Error rebuilding comments FTS index: %v", err)
		return err
	}

	log.Info.Println("FTS5 indexes rebuilt successfully")
	return nil
}

// SearchPostsFTS5 performs a fuzzy full-text search with weighted scoring
// Weights: Title (3x), Description (2x), Comments (1x)
func (ph *PostHandler) SearchPostsFTS5(query string, userID string, limit int, offset int) ([]SearchResult, error) {
	if len(query) < 3 {
		return []SearchResult{}, nil
	}

	// Prepare FTS5 query - make it fuzzy by adding wildcards
	ftsQuery := prepareFTS5Query(query)

	// Combined query with weighted scoring
	sqlQuery := `
		WITH ranked_results AS (
			-- Title matches (weight: 3)
			SELECT 
				p.id,
				'title' as match_type,
				snippet(posts_fts, 0, '<mark>', '</mark>', '...', 32) as snippet,
				rank * 3.0 as score
			FROM posts_fts
			JOIN posts p ON posts_fts.rowid = p.id
			WHERE posts_fts MATCH ?
			AND posts_fts.title MATCH ?

			UNION ALL

			-- Description matches (weight: 2)
			SELECT 
				p.id,
				'description' as match_type,
				snippet(posts_fts, 1, '<mark>', '</mark>', '...', 32) as snippet,
				rank * 2.0 as score
			FROM posts_fts
			JOIN posts p ON posts_fts.rowid = p.id
			WHERE posts_fts MATCH ?
			AND posts_fts.description MATCH ?

			UNION ALL

			-- Comment matches (weight: 1)
			SELECT 
				c.fk_post_id as id,
				'comment' as match_type,
				snippet(comments_fts, 0, '<mark>', '</mark>', '...', 32) as snippet,
				rank * 1.0 as score
			FROM comments_fts
			JOIN comments c ON comments_fts.rowid = c.id
			WHERE comments_fts MATCH ?
		)
		SELECT DISTINCT
			p.id, p.title, p.url, p.description, p.created_at, p.rank,
			u.name,
			(SELECT count(*) FROM comments c WHERE c.fk_post_id = p.id) nr_comments,
			(SELECT count(*) FROM votes v WHERE v.fk_post_id = p.id) nr_votes,
			(SELECT count(*) > 0 FROM votes v WHERE v.fk_post_id = p.id AND v.fk_user_id = ?) user_voted,
			p.fk_user_id = ? is_user_author,
			rr.match_type,
			rr.snippet,
			MAX(rr.score) as max_score
		FROM ranked_results rr
		JOIN posts p ON p.id = rr.id
		LEFT JOIN users u ON u.id = p.fk_user_id
		GROUP BY p.id
		ORDER BY max_score DESC, p.rank DESC, p.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := ph.db.Query(
		sqlQuery,
		ftsQuery, ftsQuery, // title matches
		ftsQuery, ftsQuery, // description matches
		ftsQuery,       // comment matches
		userID, userID, // user context
		limit, offset, // pagination
	)
	if err != nil {
		log.Error.Printf("Error executing FTS5 search: %v", err)
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var matchType sql.NullString
		var snippet sql.NullString
		var score sql.NullFloat64

		err := rows.Scan(
			&result.ID,
			&result.Title,
			&result.URL,
			&result.Description,
			&result.CreatedAt,
			&result.Rank,
			&result.FUserName,
			&result.FNrOfComments,
			&result.FNrOfVotes,
			&result.UserVoted,
			&result.UserIsAuthor,
			&matchType,
			&snippet,
			&score,
		)
		if err != nil {
			log.Error.Printf("Could not scan FTS5 search result: error=%v", err)
			continue
		}

		if matchType.Valid {
			result.MatchType = matchType.String
		}
		if snippet.Valid {
			result.MatchSnippet = snippet.String
		}
		if score.Valid {
			result.Score = score.Float64
		}

		results = append(results, result)
	}

	return results, nil
}

// prepareFTS5Query converts a user query into FTS5 syntax with fuzzy matching
func prepareFTS5Query(query string) string {
	// Split query into words
	words := strings.Fields(strings.TrimSpace(query))
	if len(words) == 0 {
		return ""
	}

	// Add prefix matching (fuzzy) to each word
	var ftsTerms []string
	for _, word := range words {
		// Escape special FTS5 characters
		word = strings.ReplaceAll(word, `"`, `""`)
		// Add prefix matching with *
		ftsTerms = append(ftsTerms, fmt.Sprintf(`"%s"*`, word))
	}

	// Combine with OR for fuzzy matching
	return strings.Join(ftsTerms, " OR ")
}

// CountSearchResults returns the total number of search results (for pagination)
func (ph *PostHandler) CountSearchResults(query string) (int, error) {
	if len(query) < 3 {
		return 0, nil
	}

	ftsQuery := prepareFTS5Query(query)

	sqlQuery := `
		SELECT COUNT(DISTINCT p.id)
		FROM (
			SELECT rowid as id FROM posts_fts WHERE posts_fts MATCH ?
			UNION
			SELECT fk_post_id as id FROM comments c
			JOIN comments_fts ON comments_fts.rowid = c.id
			WHERE comments_fts MATCH ?
		) matches
		JOIN posts p ON p.id = matches.id
	`

	var count int
	err := ph.db.QueryRow(sqlQuery, ftsQuery, ftsQuery).Scan(&count)
	if err != nil {
		log.Error.Printf("Error counting search results: %v", err)
		return 0, err
	}

	return count, nil
}

// SearchPostsFallback performs a LIKE-based search with weighted scoring (no FTS5 required)
// Weights: Title (3x), Description (2x), Comments (1x)
func (ph *PostHandler) SearchPostsFallback(query string, userID string, limit int, offset int) ([]SearchResult, error) {
	if len(query) < 3 {
		return []SearchResult{}, nil
	}

	// Prepare LIKE pattern for fuzzy matching
	likePattern := "%" + query + "%"

	// Combined query with weighted scoring
	sqlQuery := `
		WITH ranked_results AS (
			-- Title matches (weight: 3)
			SELECT DISTINCT
				p.id,
				'title' as match_type,
				p.title as snippet,
				3.0 as score
			FROM posts p
			WHERE p.title LIKE ?

			UNION

			-- Description matches (weight: 2)
			SELECT DISTINCT
				p.id,
				'description' as match_type,
				substr(p.description, 1, 200) as snippet,
				2.0 as score
			FROM posts p
			WHERE p.description LIKE ?

			UNION

			-- Comment matches (weight: 1)
			SELECT DISTINCT
				c.fk_post_id as id,
				'comment' as match_type,
				substr(c.text, 1, 200) as snippet,
				1.0 as score
			FROM comments c
			WHERE c.text LIKE ?
		)
		SELECT DISTINCT
			p.id, p.title, p.url, p.description, p.created_at, p.rank,
			u.name,
			(SELECT count(*) FROM comments c WHERE c.fk_post_id = p.id) nr_comments,
			(SELECT count(*) FROM votes v WHERE v.fk_post_id = p.id) nr_votes,
			(SELECT count(*) > 0 FROM votes v WHERE v.fk_post_id = p.id AND v.fk_user_id = ?) user_voted,
			p.fk_user_id = ? is_user_author,
			rr.match_type,
			rr.snippet,
			MAX(rr.score) as max_score
		FROM ranked_results rr
		JOIN posts p ON p.id = rr.id
		LEFT JOIN users u ON u.id = p.fk_user_id
		GROUP BY p.id
		ORDER BY max_score DESC, p.rank DESC, p.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := ph.db.Query(
		sqlQuery,
		likePattern,    // title matches
		likePattern,    // description matches
		likePattern,    // comment matches
		userID, userID, // user context
		limit, offset, // pagination
	)
	if err != nil {
		log.Error.Printf("Error executing fallback search: %v", err)
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var matchType sql.NullString
		var snippet sql.NullString
		var score sql.NullFloat64

		err := rows.Scan(
			&result.ID,
			&result.Title,
			&result.URL,
			&result.Description,
			&result.CreatedAt,
			&result.Rank,
			&result.FUserName,
			&result.FNrOfComments,
			&result.FNrOfVotes,
			&result.UserVoted,
			&result.UserIsAuthor,
			&matchType,
			&snippet,
			&score,
		)
		if err != nil {
			log.Error.Printf("Could not scan fallback search result: error=%v", err)
			continue
		}

		if matchType.Valid {
			result.MatchType = matchType.String
		}
		if snippet.Valid {
			// Add basic highlighting for fallback search
			result.MatchSnippet = strings.ReplaceAll(snippet.String, query, "<mark>"+query+"</mark>")
		}
		if score.Valid {
			result.Score = score.Float64
		}

		results = append(results, result)
	}

	return results, nil
}

// CountSearchResultsFallback counts results for LIKE-based search
func (ph *PostHandler) CountSearchResultsFallback(query string) (int, error) {
	if len(query) < 3 {
		return 0, nil
	}

	likePattern := "%" + query + "%"

	sqlQuery := `
		SELECT COUNT(DISTINCT p.id)
		FROM posts p
		LEFT JOIN comments c ON c.fk_post_id = p.id
		WHERE p.title LIKE ? OR p.description LIKE ? OR c.text LIKE ?
	`

	var count int
	err := ph.db.QueryRow(sqlQuery, likePattern, likePattern, likePattern).Scan(&count)
	if err != nil {
		log.Error.Printf("Error counting fallback search results: %v", err)
		return 0, err
	}

	return count, nil
}
