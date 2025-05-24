import { Router } from 'express'
import { AuthController, PostController, CommentController } from '../controllers'

const router = Router()

// Authentication routes
router.get('/login', (req, res) => res.render('login'))
router.post('/login', AuthController.login)
router.get('/logout', AuthController.logout)
router.get('/register', (req, res) => res.render('register'))
router.post('/register', AuthController.register)

// Post routes
router.get('/', PostController.index)
router.get('/urls', PostController.urls)
router.get('/questions', PostController.questions)
router.get('/post/new', PostController.new)
router.post('/post', PostController.create)
router.get('/post/:id', PostController.show)
router.post('/post/:id/delete', PostController.delete)
router.post('/post/:id/vote', PostController.vote)

// Comment routes
router.post('/post/:postId/comment', CommentController.create)
router.post('/comment/:id/delete', CommentController.delete)
router.post('/comment/:id/update', CommentController.update)
router.post('/comment/:id/vote', CommentController.vote)

export default router
