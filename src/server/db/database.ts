import Database from 'better-sqlite3'
import fs from 'fs'
import path from 'path'

const dbPath = path.join(__dirname, 'agora.db')
const db = new Database(dbPath)

// Create database tables if they don't exist
function initializeDatabase() {
	// Create users table
	db.exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)

	// Create posts table (for both URLs and questions)
	db.exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			url TEXT,
			description TEXT,
			type TEXT NOT NULL, -- 'url' or 'question'
			score INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)
	`)

	// Create comments table
	db.exec(`
		CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			parent_id INTEGER, -- NULL for top-level comments
			content TEXT NOT NULL,
			score INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(post_id) REFERENCES posts(id),
			FOREIGN KEY(user_id) REFERENCES users(id),
			FOREIGN KEY(parent_id) REFERENCES comments(id)
		)
	`)

	// Create votes table
	db.exec(`
		CREATE TABLE IF NOT EXISTS votes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			post_id INTEGER,
			comment_id INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id),
			FOREIGN KEY(post_id) REFERENCES posts(id),
			FOREIGN KEY(comment_id) REFERENCES comments(id),
			CHECK ((post_id IS NULL AND comment_id IS NOT NULL) OR (post_id IS NOT NULL AND comment_id IS NULL))
		)
	`)

	// Add some initial data for testing
	const userCount = db.prepare('SELECT COUNT(*) as count FROM users').get()
	
	if (userCount.count === 0) {
		// Add a test user
		db.prepare('INSERT INTO users (username, password_hash) VALUES (?, ?)').run('testuser', 'password123')
		
		// Add some test posts
		db.prepare('INSERT INTO posts (user_id, title, url, description, type) VALUES (?, ?, ?, ?, ?)').run(
			1, 
			'Welcome to Agora', 
			null, 
			'This is a discussion platform similar to Hacker News',
			'question'
		)
		
		db.prepare('INSERT INTO posts (user_id, title, url, description, type) VALUES (?, ?, ?, ?, ?)').run(
			1, 
			'GitHub - The world\'s leading developer platform', 
			'https://github.com', 
			'Where developers build, ship, and maintain their software',
			'url'
		)
	}
}

// Initialize the database
initializeDatabase()

export default db
