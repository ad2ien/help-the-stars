package main

import (
	"embed"
	"flag"
	"help-the-stars/internal"
	"net/http"
	"os"

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
	webpageHandler := internal.CreateWebpageHandler(controller, &templates)

	http.HandleFunc("/", webpageHandler.HandleWebPage)

	log.Info("Server listening on port 1983")
	err := http.ListenAndServe(":1983", nil)

	if err != nil {
		log.Error("Server error", err)
	}
}
