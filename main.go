package main

import (
	"embed"
	"flag"
	"fmt"
	"help-the-stars/internal"
	"html/template"
	"log"
	"net/http"
	"os"
)

var version = "1.0.0"

//go:embed migrations
var Migrations embed.FS

//go:embed templates
var templates embed.FS

func main() {

	helpFlag := flag.Bool("help", false, "Display help information")
	versionFlag := flag.Bool("version", false, "Display version information")

	flag.Parse()

	if *helpFlag {
		fmt.Println("Usage of Help the stars:")
		fmt.Println("  -help\tDisplay help information")
		fmt.Println("  -version\tDisplay version information")
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Println("Version:", version)
		os.Exit(0)
	}

	internal.GetSettings()

	matrix := internal.CreateMatrixClient()

	db := internal.NewConnection(Migrations)
	defer db.Close()

	controller := internal.CreateController(db.Connection, matrix)

	fmt.Println("Start worker...")
	go controller.Worker()

	fmt.Println("starting server...")
	startServer(controller)
}

func startServer(controller *internal.DataController) {

	tmpl := template.Must(template.ParseFS(templates, "templates/index.html"))

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

	log.Println("Server listening on port 1983")
	http.ListenAndServe(":1983", nil)
}
