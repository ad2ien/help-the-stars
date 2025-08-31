package internal

import (
	"embed"
	"net/http"
	"text/template"
	"time"

	"github.com/charmbracelet/log"
)

type WebpageHandler struct {
	dataController *DataController
	templates      *embed.FS
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func uniqueRepoOwner(issue []HelpWantedIssue) []string {
	repoOwners := make(map[string]bool)
	var uniqueRepoOwners []string
	for _, issue := range issue {
		if _, ok := repoOwners[issue.RepoOwner]; !ok {
			repoOwners[issue.RepoOwner] = true
			uniqueRepoOwners = append(uniqueRepoOwners, issue.RepoOwner)
		}
	}
	return uniqueRepoOwners
}

func CreateWebpageHandler(dataController *DataController, templates *embed.FS) *WebpageHandler {
	return &WebpageHandler{
		dataController: dataController,
		templates:      templates,
	}
}

func (wph *WebpageHandler) HandleWebPage(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.New("index.html").Funcs(template.FuncMap{
		"date":            formatDate,
		"uniqueRepoOwner": uniqueRepoOwner,
	}).ParseFS(wph.templates, "templates/index.html"))

	data, err := wph.dataController.GetDataForView()
	if err != nil {
		log.Error("Error getting data:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Error("Error executing template:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
