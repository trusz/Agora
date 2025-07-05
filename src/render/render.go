package render

import (
	"agora/src/log"
	"agora/src/server/auth"
	"context"
	_ "embed"
	"net/http"
	"text/template"
)

//go:embed layout.html
var layoutTemplate string

//go:embed header.html
var headerTemplate string

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

	var templates []string

	templates = append(templates, layoutTemplate)
	templates = append(templates, headerTemplate)
	templates = append(templates, includedTemplates...)

	// Parse all templates
	parsedTemplates := template.New("all")
	for _, tmpl := range templates {
		var err error
		parsedTemplates, err = parsedTemplates.Parse(tmpl)
		if err != nil {
			log.Error.Printf("msg='could not parse template' template='%s' err='%s'\n", tmpl, err.Error())
		}
	}

	page.User.Name = user.Name

	err := parsedTemplates.ExecuteTemplate(w, templateToExecute, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
