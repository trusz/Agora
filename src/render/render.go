package render

import (
	"net/http"
	"path/filepath"
	"text/template"
)

func RenderTemplate(w http.ResponseWriter, templateToExecute string, p *Page, includedTemplates ...string) {
	renderDir := "src/render"
	templates, err := filepath.Glob(renderDir + "/*.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add the specific template if it's not already included
	alreadyIncluded := false
	// for _, tmpl := range tmpls {
	for _, t := range templates {
		if t == templateToExecute {
			alreadyIncluded = true
			break
		}
	}
	if !alreadyIncluded {
		templates = append(templates, templateToExecute)
	}
	// }
	templates = append(templates, includedTemplates...)

	// Parse all templates
	ts, err := template.ParseFiles(templates...)
	if err != nil {
		http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, filepath.Base(templateToExecute), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
