package render

import (
	"goazuread/src/log"
	"net/http"
	"path/filepath"
	"text/template"
)

var ts *template.Template

// Run this function on startup
func PreRenderAllHTML() {
	templates, err := filepath.Glob("src/**/*.html")
	if err != nil {
		log.Error.Printf("msg='finding tempaltes failed, stopping', err='%s'", err)
		return
	}
	log.Debug.Println("templates", templates)
	ts, err = template.ParseFiles(templates...)
	if err != nil {
		log.Error.Printf("msg='could not parse all templates, stopping' err='%s'", err)

	}
}

func RenderTemplateV2(w http.ResponseWriter, tmpl string, p *Page) {
	log.Debug.Println("Rendering with V2", tmpl)
	if ts == nil {
		log.Error.Printf("msg='no rendered templates, stopping'")
		http.Error(w, "no prerendered template", http.StatusInternalServerError)
		return
	}
	// Execute the template file specified by tmpl
	log.Debug.Println("filepathbase", filepath.Base(tmpl), ts)
	err := ts.ExecuteTemplate(w, filepath.Base(tmpl), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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
	ts1, err := template.ParseFiles(templates...)
	if err != nil {
		http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = ts1.ExecuteTemplate(w, filepath.Base(tmpl), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
