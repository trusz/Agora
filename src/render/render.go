package render

import (
	"net/http"
	"path/filepath"
	"text/template"
)

func RenderTemplate(w http.ResponseWriter, tmpl string, p *Page) {

	layoutFile := "src/render/layout.html"
	ts, err := template.ParseFiles(layoutFile, tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the template file specified by tmpl
	err = ts.ExecuteTemplate(w, filepath.Base(tmpl), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
