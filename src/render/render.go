package render

import (
	"agora/src/log"
	"agora/src/server/auth"
	"context"
	"net/http"
	"path/filepath"
	"text/template"
)

func RenderTemplate(
	w http.ResponseWriter,
	templateToExecute string,
	page *Page,
	ctx context.Context,
	includedTemplates ...string,
) {
	user, ok := auth.ExtractUserFromContext(ctx)
	if !ok {
		log.Error.Printf("msg='could not get user from context' context='%#v'\n", ctx)
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	renderDir := "src/render"
	templates, err := filepath.Glob(renderDir + "/*.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add the specific template if it's not already included
	alreadyIncluded := false

	for _, t := range templates {
		if t == templateToExecute {
			alreadyIncluded = true
			break
		}
	}
	if !alreadyIncluded {
		templates = append(templates, templateToExecute)
	}

	templates = append(templates, includedTemplates...)

	// Parse all templates
	ts, err := template.ParseFiles(templates...)
	if err != nil {
		http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	page.User.Name = user.Name

	err = ts.ExecuteTemplate(w, filepath.Base(templateToExecute), page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
