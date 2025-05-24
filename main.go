package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
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
	// Define file server for static files
	fs := http.FileServer(http.Dir("./src/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Define routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/settings", settingsHandler)

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
