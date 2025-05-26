package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"Agora/src/auth"
	_ "github.com/mattn/go-sqlite3"
)

// Page represents data to be displayed in our templates
type Page struct {
	Title       string
	Data        interface{}
	CurrentYear int
}

// Post represents a blog post
type Post struct {
	Title   string
	Content string
	Author  string
}

// Settings represents user settings
type Settings struct {
	Theme                string
	ShowAvatar           bool
	NewsletterSubscribed bool
}

const defaultPort = ":57694"

func main() {
	// Initialize SQLite database
	db, err := sql.Open("sqlite3", "./src/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create required tables
	if err := initDatabase(db); err != nil {
		log.Fatal(err)
	}

	// Load authentication configuration
	authConfig := auth.Config{
		TenantID:     getEnvOrDefault("AZURE_TENANT_ID", "common"),
		ClientID:     getEnvOrDefault("AZURE_CLIENT_ID", "your-client-id"),
		ClientSecret: getEnvOrDefault("AZURE_CLIENT_SECRET", "your-client-secret"),
		RedirectURI:  getEnvOrDefault("AZURE_REDIRECT_URI", "http://localhost"+defaultPort+"/auth/callback"),
	}

	// Initialize auth service
	authService := auth.NewAuth(db, authConfig)

	// Define file server for static files
	fs := http.FileServer(http.Dir("./src/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Define auth routes
	http.HandleFunc("/login", authService.LoginHandler)
	http.HandleFunc("/logout", authService.LogoutHandler)
	http.HandleFunc("/auth/callback", authService.CallbackHandler)

	// Define public routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)

	// Protected routes (require authentication)
	http.Handle("/post", authService.AuthMiddleware(http.HandlerFunc(postHandler)))
	http.Handle("/settings", authService.AuthMiddleware(http.HandlerFunc(settingsHandler)))

	// Start server
	log.Printf("Starting server on %s", defaultPort)
	log.Fatal(http.ListenAndServe(defaultPort, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect to about page if accessing just the root
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, "/about", http.StatusSeeOther)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "about", &Page{
		Title: "About",
		Data:  nil,
	})
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	// Sample post data
	post := Post{
		Title:   "Introduction to Go Templates",
		Content: "Go templates are a powerful way to generate HTML content dynamically. This system allows you to create reusable templates for your web applications.",
		Author:  "Go Developer",
	}

	renderTemplate(w, "post", &Page{
		Title: post.Title,
		Data:  post,
	})
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	// Sample settings data
	settings := Settings{
		Theme:                "light",
		ShowAvatar:           true,
		NewsletterSubscribed: false,
	}

	renderTemplate(w, "settings", &Page{
		Title: "Settings",
		Data:  settings,
	})
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	// Add current year for the footer
	p.CurrentYear = time.Now().Year()

	// Get all template files
	templates, err := filepath.Glob("src/templates/*.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a template set with all templates
	ts, err := template.ParseFiles(templates...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the named template
	err = ts.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// initDatabase initializes the database schema
func initDatabase(db *sql.DB) error {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create sessions table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
