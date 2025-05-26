package render

import (
	"net/http"
	"path/filepath"
	"text/template"
)

func RenderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	renderDir := "src/render"
	templates, err := filepath.Glob(renderDir + "/*.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add the specific template if it's not already included
	alreadyIncluded := false
	for _, t := range templates {
		if t == tmpl {
			alreadyIncluded = true
			break
		}
	}

	if !alreadyIncluded {
		templates = append(templates, tmpl)
	}

	// Parse all templates
	ts, err := template.ParseFiles(templates...)
	if err != nil {
		http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the template file specified by tmpl
	err = ts.ExecuteTemplate(w, filepath.Base(tmpl), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
