package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

var version = "1.0.0"

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

	GetSettings()

	fmt.Println("Loading issues...")
	data, err := GetStaredRepos(50)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("data loaded, starting server...")
	startServer(data)
}

func startServer(data []HelpWantedIssue) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("Server listening on port 1983")
	http.ListenAndServe(":1983", nil)
}
