import db from '../db/database'

// Types
export interface User {
	id: number
	username: string
	created_at: string
}

export interface Post {
	id: number
	user_id: number
	title: string
	url: string | null
	description: string | null
	type: 'url' | 'question'
	score: number
	created_at: string
	username?: string // Joined field
	comments_count?: number // Calculated field
}

export interface Comment {
	id: number
	post_id: number
	user_id: number
	parent_id: number | null
	content: string
	score: number
	created_at: string
	username?: string // Joined field
}

export interface Vote {
	id: number
	user_id: number
	post_id: number | null
	comment_id: number | null
	created_at: string
}

// User model functions
export const UserModel = {
	findById: (id: number): User | undefined => {
		return db.prepare('SELECT id, username, created_at FROM users WHERE id = ?').get(id) as User | undefined
	},
	
	findByUsername: (username: string): User | undefined => {
		return db.prepare('SELECT id, username, created_at FROM users WHERE username = ?').get(username) as User | undefined
	},
	
	create: (username: string, passwordHash: string): User => {
		const result = db.prepare('INSERT INTO users (username, password_hash) VALUES (?, ?) RETURNING id, username, created_at').get(username, passwordHash) as User
		return result
	}
}

// Post model functions
export const PostModel = {
	findById: (id: number): Post | undefined => {
		return db.prepare(`
			SELECT posts.*, users.username 
			FROM posts 
			JOIN users ON posts.user_id = users.id 
			WHERE posts.id = ?
		`).get(id) as Post | undefined
	},
	
	findAll: (type?: 'url' | 'question', limit = 30, offset = 0): Post[] => {
		const query = `
			SELECT 
				posts.*, 
				users.username,
				(SELECT COUNT(*) FROM comments WHERE post_id = posts.id) as comments_count
			FROM posts 
			JOIN users ON posts.user_id = users.id
			${type ? 'WHERE posts.type = ?' : ''}
			ORDER BY posts.score DESC, posts.created_at DESC
			LIMIT ? OFFSET ?
		`
		
		return type 
			? db.prepare(query).all(type, limit, offset) as Post[]
			: db.prepare(query).all(limit, offset) as Post[]
	},
	
	create: (post: Omit<Post, 'id' | 'score' | 'created_at'>): Post => {
		const result = db.prepare(`
			INSERT INTO posts (user_id, title, url, description, type) 
			VALUES (?, ?, ?, ?, ?) 
			RETURNING *
		`).get(post.user_id, post.title, post.url, post.description, post.type) as Post
		
		return result
	},
	
	delete: (id: number, userId: number): boolean => {
		// Only allow the user who created the post to delete it
		const result = db.prepare('DELETE FROM posts WHERE id = ? AND user_id = ?').run(id, userId)
		return result.changes > 0
	},
	
	vote: (postId: number, userId: number): number => {
		// Check if user already voted for this post
		const existingVote = db.prepare('SELECT id FROM votes WHERE post_id = ? AND user_id = ?').get(postId, userId)
		
		// If the user already voted, remove their vote
		if (existingVote) {
			db.prepare('DELETE FROM votes WHERE id = ?').run(existingVote.id)
			db.prepare('UPDATE posts SET score = score - 1 WHERE id = ?').run(postId)
			return -1
		}
		
		// Otherwise, add a new vote
		db.prepare('INSERT INTO votes (user_id, post_id) VALUES (?, ?)').run(userId, postId)
		db.prepare('UPDATE posts SET score = score + 1 WHERE id = ?').run(postId)
		return 1
	}
}

// Comment model functions
export const CommentModel = {
	findById: (id: number): Comment | undefined => {
		return db.prepare(`
			SELECT comments.*, users.username 
			FROM comments 
			JOIN users ON comments.user_id = users.id 
			WHERE comments.id = ?
		`).get(id) as Comment | undefined
	},
	
	findByPostId: (postId: number): Comment[] => {
		return db.prepare(`
			SELECT comments.*, users.username 
			FROM comments 
			JOIN users ON comments.user_id = users.id 
			WHERE comments.post_id = ? 
			ORDER BY comments.score DESC, comments.created_at ASC
		`).all(postId) as Comment[]
	},
	
	create: (comment: Omit<Comment, 'id' | 'score' | 'created_at'>): Comment => {
		const result = db.prepare(`
			INSERT INTO comments (post_id, user_id, parent_id, content) 
			VALUES (?, ?, ?, ?) 
			RETURNING *
		`).get(comment.post_id, comment.user_id, comment.parent_id, comment.content) as Comment
		
		return result
	},
	
	delete: (id: number, userId: number): boolean => {
		// Only allow the user who created the comment to delete it
		const result = db.prepare('DELETE FROM comments WHERE id = ? AND user_id = ?').run(id, userId)
		return result.changes > 0
	},
	
	update: (id: number, userId: number, content: string): Comment | undefined => {
		// Only allow the user who created the comment to update it
		const result = db.prepare(`
			UPDATE comments SET content = ? WHERE id = ? AND user_id = ? RETURNING *
		`).get(content, id, userId) as Comment | undefined
		
		return result
	},
	
	vote: (commentId: number, userId: number): number => {
		// Check if user already voted for this comment
		const existingVote = db.prepare('SELECT id FROM votes WHERE comment_id = ? AND user_id = ?').get(commentId, userId)
		
		// If the user already voted, remove their vote
		if (existingVote) {
			db.prepare('DELETE FROM votes WHERE id = ?').run(existingVote.id)
			db.prepare('UPDATE comments SET score = score - 1 WHERE id = ?').run(commentId)
			return -1
		}
		
		// Otherwise, add a new vote
		db.prepare('INSERT INTO votes (user_id, comment_id) VALUES (?, ?)').run(userId, commentId)
		db.prepare('UPDATE comments SET score = score + 1 WHERE id = ?').run(commentId)
		return 1
	}
}
