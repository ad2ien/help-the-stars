package main

import (
	"embed"
	"flag"
	"fmt"
	"help-the-stars/internal"

	"github.com/charmbracelet/log"
)

//go:embed migrations
var Migrations embed.FS

//go:embed templates
var templates embed.FS

//go:embed static/*
var staticFiles embed.FS

func main() {
	flag.Usage = func() {
		fmt.Println("Usage of Help the start ⭐")
		fmt.Println("  https://github.com/ad2ien/help-the-stars/")
		fmt.Println(".env example :")
		fmt.Println("\tMATRIX_TOKEN=your-matrix-token")
		fmt.Println("\tMATRIX_USERID=your-matrix-userid")
		fmt.Println("\tMATRIX_ROOMID=your-matrix-roomid")
		fmt.Println("\tDB_FILE=db/help-the-stars-dev.db")
		fmt.Println("\tLABELS='\"help-wanted\",\"junior friendly\",\"good first issue\"'")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
	}

	debugFlag := flag.Bool("debug", false, "Display debug logs")

	flag.Parse()

	internal.GetSettings()

	matrix := internal.CreateMatrixClient()

	db := internal.NewConnection(Migrations)
	defer db.Close()

	controller := internal.CreateController(db.Connection, matrix)

	log.Info("Start worker...")
	if *debugFlag {
		log.SetLevel(log.DebugLevel)
	}
	log.Debug("Debugging on")
	go controller.Worker()

	server := internal.NewServer(controller)
	server.Start(&templates, &staticFiles)
}
