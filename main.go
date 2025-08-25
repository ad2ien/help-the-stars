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

	db := internal.NewConnection(Migrations)
	defer db.Close()

	controller := internal.CreateController(db.Connection)

	fmt.Println("Start worker...")
	go controller.Worker()

	fmt.Println("starting server...")
	startServer()
}

func startServer() {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// http.HandleFunc("/thanksdata", func(w http.ResponseWriter, r *http.Request) {
	// 	data := internal.GetNextPage("")
	// 	json.NewEncoder(w).Encode(data)
	// })

	log.Println("Server listening on port 1983")
	http.ListenAndServe(":1983", nil)
}
