package internal

import (
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/log"
)

type WebpageHandler struct {
	dataController *DataController
	templates      *embed.FS
}

func CreateWebpageHandler(dataController *DataController, templates *embed.FS) *WebpageHandler {
	return &WebpageHandler{
		dataController: dataController,
		templates:      templates,
	}
}

func (wph *WebpageHandler) HandleWebPage(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.New("index.html").Funcs(template.FuncMap{
		"truncate":            truncate,
		"date":                formatDate,
		"buildHelpIssuesLink": buildHelpIssuesLink,
	}).ParseFS(wph.templates, "templates/index.html"))

	data, err := wph.dataController.GetDataForView()
	if err != nil {
		log.Error("Error getting data:", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Error("Error executing template:", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

func truncate(s string, length int) string {
	if len(s) > length {
		return s[:length] + "..."
	}
	return s
}

const ISSUE_LINK_PARAM = "issues?q=is%3Aissue%20state%3Aopen%20"

func buildHelpIssuesLink(repoOwner string) string {
	return fmt.Sprintf("https://github.com/%s/%s%s", repoOwner, ISSUE_LINK_PARAM, labelsToGhUrlParam())
}

// TransformLabels transforms a string like `"good first issue", "help wanted"`
// into `(label%3A%22good%20first%20issue%22%20OR%20label%3A%22help%20wanted%22)`.
// to have something like https://github.com/OWNER/REPO/issues?q=is"issue state=open (label="good first issue" OR label="help wanted")
func labelsToGhUrlParam() string {
	labelSettings := GetSettings().GetLabelSlice()

	var labels []string
	for _, label := range labelSettings {
		// URL encode the label (replace spaces with %20)
		encodedLabel := strings.ReplaceAll(url.QueryEscape(label), "+", "%20")
		labels = append(labels, fmt.Sprintf("label:%%22%s%%22", encodedLabel))
	}

	// Join labels with " OR " and wrap in parentheses
	if len(labels) > 0 {
		return fmt.Sprintf("(%s)", strings.Join(labels, "%20OR%20"))
	}
	return ""
}
