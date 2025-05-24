import { Request, Response } from 'express'
import { PostModel, UserModel, CommentModel } from '../models'

// Auth controller (simplified for demo)
export const AuthController = {
	login: (req: Request, res: Response) => {
		const { username, password } = req.body
		
		if (!username || !password) {
			return res.status(400).render('error', { message: 'Username and password are required' })
		}
		
		// In a real app, we would check password hash
		const user = UserModel.findByUsername(username)
		
		if (!user) {
			// For demo purposes, create a user if it doesn't exist
			const newUser = UserModel.create(username, 'password123')
			req.session.userId = newUser.id
			return res.redirect('/')
		}
		
		req.session.userId = user.id
		return res.redirect('/')
	},
	
	logout: (req: Request, res: Response) => {
		req.session.userId = undefined
		res.redirect('/')
	},
	
	register: (req: Request, res: Response) => {
		const { username, password } = req.body
		
		if (!username || !password) {
			return res.status(400).render('error', { message: 'Username and password are required' })
		}
		
		// Check if user already exists
		const existingUser = UserModel.findByUsername(username)
		
		if (existingUser) {
			return res.status(400).render('error', { message: 'Username already taken' })
		}
		
		// In a real app, we would hash the password
		const newUser = UserModel.create(username, 'password123')
		req.session.userId = newUser.id
		
		return res.redirect('/')
	}
}

// Posts controller
export const PostController = {
	// Get all posts (homepage)
	index: (req: Request, res: Response) => {
		const posts = PostModel.findAll()
		res.render('index', { 
			posts,
			user: req.session.userId ? UserModel.findById(req.session.userId) : null
		})
	},
	
	// Get posts of type 'url'
	urls: (req: Request, res: Response) => {
		const posts = PostModel.findAll('url')
		res.render('index', { 
			posts,
			postType: 'url',
			user: req.session.userId ? UserModel.findById(req.session.userId) : null
		})
	},
	
	// Get posts of type 'question'
	questions: (req: Request, res: Response) => {
		const posts = PostModel.findAll('question')
		res.render('index', { 
			posts,
			postType: 'question',
			user: req.session.userId ? UserModel.findById(req.session.userId) : null
		})
	},
	
	// Show post form
	new: (req: Request, res: Response) => {
		// Ensure user is logged in
		if (!req.session.userId) {
			return res.redirect('/login')
		}
		
		res.render('post-form', { 
			user: UserModel.findById(req.session.userId),
			postType: req.query.type || 'url' 
		})
	},
	
	// Create a new post
	create: (req: Request, res: Response) => {
		// Ensure user is logged in
		if (!req.session.userId) {
			return res.status(401).render('error', { message: 'You must be logged in to create a post' })
		}
		
		const { title, url, description, type } = req.body
		
		if (!title || !type) {
			return res.status(400).render('error', { message: 'Title and type are required' })
		}
		
		// Validate URL for URL-type posts
		if (type === 'url' && !url) {
			return res.status(400).render('error', { message: 'URL is required for URL posts' })
		}
		
		const newPost = PostModel.create({
			user_id: req.session.userId,
			title,
			url: type === 'url' ? url : null,
			description,
			type
		})
		
		return res.redirect(`/post/${newPost.id}`)
	},
	
	// Show a single post with comments
	show: (req: Request, res: Response) => {
		const postId = parseInt(req.params.id)
		const post = PostModel.findById(postId)
		
		if (!post) {
			return res.status(404).render('error', { message: 'Post not found' })
		}
		
		const comments = CommentModel.findByPostId(postId)
		
		// Organize comments into a tree structure
		const commentTree: Record<number, any[]> = { 0: [] } // Root level is 0
		
		comments.forEach(comment => {
			// Initialize an array for this comment's replies if it doesn't exist
			if (!commentTree[comment.id]) {
				commentTree[comment.id] = []
			}
			
			// Add this comment to its parent's array
			const parentId = comment.parent_id || 0
			if (!commentTree[parentId]) {
				commentTree[parentId] = []
			}
			
			commentTree[parentId].push({
				...comment,
				replies: commentTree[comment.id]
			})
		})
		
		res.render('post', { 
			post,
			comments: commentTree[0], // Root level comments
			user: req.session.userId ? UserModel.findById(req.session.userId) : null
		})
	},
	
	// Delete a post
	delete: (req: Request, res: Response) => {
		// Ensure user is logged in
		if (!req.session.userId) {
			return res.status(401).json({ error: 'You must be logged in to delete a post' })
		}
		
		const postId = parseInt(req.params.id)
		const deleted = PostModel.delete(postId, req.session.userId)
		
		if (!deleted) {
			return res.status(403).json({ error: 'You can only delete your own posts' })
		}
		
		return res.redirect('/')
	},
	
	// Vote for a post
	vote: (req: Request, res: Response) => {
		// Ensure user is logged in
		if (!req.session.userId) {
			return res.status(401).json({ error: 'You must be logged in to vote' })
		}
		
		const postId = parseInt(req.params.id)
		const result = PostModel.vote(postId, req.session.userId)
		
		return res.json({ success: true, score_change: result })
	}
}

// Comments controller
export const CommentController = {
	// Create a new comment
	create: (req: Request, res: Response) => {
		// Ensure user is logged in
		if (!req.session.userId) {
			return res.status(401).json({ error: 'You must be logged in to comment' })
		}
		
		const postId = parseInt(req.params.postId)
		const { content, parent_id } = req.body
		
		if (!content) {
			return res.status(400).json({ error: 'Comment content is required' })
		}
		
		const newComment = CommentModel.create({
			post_id: postId,
			user_id: req.session.userId,
			parent_id: parent_id ? parseInt(parent_id) : null,
			content
		})
		
		// For API requests
		if (req.xhr) {
			return res.json({ 
				success: true, 
				comment: {
					...newComment,
					username: UserModel.findById(req.session.userId)?.username
				} 
			})
		}
		
		// For form submissions
		return res.redirect(`/post/${postId}`)
	},
	
	// Delete a comment
	delete: (req: Request, res: Response) => {
		// Ensure user is logged in
		if (!req.session.userId) {
			return res.status(401).json({ error: 'You must be logged in to delete a comment' })
		}
		
		const commentId = parseInt(req.params.id)
		const comment = CommentModel.findById(commentId)
		
		if (!comment) {
			return res.status(404).json({ error: 'Comment not found' })
		}
		
		const deleted = CommentModel.delete(commentId, req.session.userId)
		
		if (!deleted) {
			return res.status(403).json({ error: 'You can only delete your own comments' })
		}
		
		return res.json({ success: true })
	},
	
	// Update a comment
	update: (req: Request, res: Response) => {
		// Ensure user is logged in
		if (!req.session.userId) {
			return res.status(401).json({ error: 'You must be logged in to edit a comment' })
		}
		
		const commentId = parseInt(req.params.id)
		const { content } = req.body
		
		if (!content) {
			return res.status(400).json({ error: 'Comment content is required' })
		}
		
		const updatedComment = CommentModel.update(commentId, req.session.userId, content)
		
		if (!updatedComment) {
			return res.status(403).json({ error: 'You can only edit your own comments' })
		}
		
		return res.json({ 
			success: true, 
			comment: {
				...updatedComment,
				username: UserModel.findById(req.session.userId)?.username
			} 
		})
	},
	
	// Vote for a comment
	vote: (req: Request, res: Response) => {
		// Ensure user is logged in
		if (!req.session.userId) {
			return res.status(401).json({ error: 'You must be logged in to vote' })
		}
		
		const commentId = parseInt(req.params.id)
		const result = CommentModel.vote(commentId, req.session.userId)
		
		return res.json({ success: true, score_change: result })
	}
}
