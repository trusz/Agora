import express from 'express'
import path from 'node:path'
import bodyParser from 'body-parser'
import routes from './routes'
import * as crypto from 'node:crypto'
import * as fs from 'node:fs'

// Express app setup
const app = express()
const port = process.env.PORT || 51872

// Simple cookie parser
interface Cookies {
  [key: string]: string
}

// Extend Request type
declare global {
  namespace Express {
    interface Request {
      session: {
        userId?: number
      }
      xhr?: boolean
      cookies?: Cookies
    }
  }
}

// A simple session implementation for this demo
const SESSIONS_FILE = path.join(__dirname, 'sessions.json')
const sessionStore: Record<string, {userId?: number, expires: number}> = {}

// Load existing sessions if available
try {
  if (fs.existsSync(SESSIONS_FILE)) {
    const data = fs.readFileSync(SESSIONS_FILE, 'utf8')
    Object.assign(sessionStore, JSON.parse(data))
    
    // Clean up expired sessions
    const now = Date.now()
    for (const [id, session] of Object.entries(sessionStore)) {
      if (session.expires < now) {
        delete sessionStore[id]
      }
    }
  }
} catch (error) {
  console.error('Error loading sessions:', error)
}

// Middleware
app.use(bodyParser.urlencoded({ extended: true }))
app.use(bodyParser.json())
app.use(express.static(path.join(__dirname, '../client/public')))

// Simple cookie parser middleware
app.use((req, res, next) => {
  const cookies: Cookies = {}
  const cookieHeader = req.headers.cookie
  
  if (cookieHeader) {
    for (const cookie of cookieHeader.split(';')) {
      const parts = cookie.split('=')
      const key = parts[0].trim()
      const value = parts[1]?.trim() || ''
      cookies[key] = value
    }
  }
  
  req.cookies = cookies
  next()
})

// Session middleware
app.use((req, res, next) => {
  let sessionId = req.cookies?.sessionId
  
  if (!sessionId || !sessionStore[sessionId]) {
    // Create new session
    sessionId = crypto.randomBytes(32).toString('hex')
    res.setHeader('Set-Cookie', `sessionId=${sessionId}; HttpOnly; Max-Age=${7 * 24 * 60 * 60}; Path=/`)
    
    sessionStore[sessionId] = {
      expires: Date.now() + (7 * 24 * 60 * 60 * 1000)
    }
  }
  
  // Extend session expiry
  sessionStore[sessionId].expires = Date.now() + (7 * 24 * 60 * 60 * 1000)
  
  // Add session data to request
  req.session = sessionStore[sessionId]
  
  // Save sessions periodically
  if (Math.random() < 0.1) {
    fs.writeFileSync(SESSIONS_FILE, JSON.stringify(sessionStore))
  }
  
  // Check for XHR requests
  req.xhr = req.headers['x-requested-with'] === 'XMLHttpRequest'
  
  next()
})

// Set view engine
app.set('view engine', 'ejs')
app.set('views', path.join(__dirname, '../client/views'))

// Routes
app.use('/', routes)

// Error handler
app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error(err.stack)
  res.status(500).render('error', { message: 'Something went wrong!' })
})

// Start server
app.listen(port, () => {
  console.log(`Agora app listening at http://localhost:${port}`)
})

export default app

// Set view engine
app.set('view engine', 'ejs')
app.set('views', path.join(__dirname, '../client/views'))

// Routes
app.use('/', routes)

// Error handler
app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
	console.error(err.stack)
	res.status(500).render('error', { message: 'Something went wrong!' })
})

// Start server
app.listen(port, () => {
	console.log(`Agora app listening at http://localhost:${port}`)
})

export default app
