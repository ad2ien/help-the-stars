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
		fmt.Println("Usage of Help the stars ⭐")
		fmt.Println("  https://github.com/ad2ien/help-the-stars/")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
	}

	debugFlag := flag.Bool("debug", false, "Display debug logs")
	interval := flag.Int("interval", 7, "Hours interval between github queries")
	ghTokenFlag := flag.String("gh-token", "", "Github token")
	labels := flag.String("labels", "\"help-wanted\", \"help wanted\",\"junior friendly\",\"good first issue\"", "labels to look for")
	dbFile := flag.String("db-file", "db/help-the-stars.db", "SQLite database file")
	matrixServer := flag.String("matrix-server", "", "Matrix homeserver URL")
	matrixUsername := flag.String("matrix-user", "", "Matrix user")
	matrixPassword := flag.String("matrix-password", "", "Matrix password")
	matrixRoomID := flag.String("matrix-room", "", "Matrix room ID")

	flag.Parse()

	internal.SetSettings(&internal.Settings{
		GhToken:        *ghTokenFlag,
		Interval:       *interval,
		Labels:         *labels,
		DBFile:         *dbFile,
		MatrixServer:   *matrixServer,
		MatrixUsername: *matrixUsername,
		MatrixPassword: *matrixPassword,
		MatrixRoomID:   *matrixRoomID,
	})

	internal.GetSettings().Print()

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
