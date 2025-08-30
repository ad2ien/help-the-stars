package main

import (
	"embed"
	"flag"
	"help-the-stars/internal"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/log"
)

//go:embed migrations
var Migrations embed.FS

//go:embed templates
var templates embed.FS

func main() {

	helpFlag := flag.Bool("help", false, "Display help information")

	flag.Parse()

	if *helpFlag {
		log.Info("Usage of Help the stars:")
		log.Info("  - https://github.com/ad2ien/help-the-stars/")
		os.Exit(0)
	}

	internal.GetSettings()

	matrix := internal.CreateMatrixClient()

	db := internal.NewConnection(Migrations)
	defer db.Close()

	controller := internal.CreateController(db.Connection, matrix)

	log.Info("Start worker...")
	go controller.Worker()

	log.Info("Start server...")
	startServer(controller)
}

func startServer(controller *internal.DataController) {

	tmpl := template.Must(template.New("index.html").Funcs(template.FuncMap{
		"date": formatDate,
	}).ParseFS(templates, "templates/index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		data, err := controller.GetDataForView()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			err := tmpl.Execute(w, data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})

	log.Info("Server listening on port 1983")
	http.ListenAndServe(":1983", nil)
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
