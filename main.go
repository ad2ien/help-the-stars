package main

import (
	"context"
	"embed"
	"flag"
	"help-the-stars/internal"

	"github.com/charmbracelet/log"
)

const defaultInterfalHours = 7

//go:embed migrations
var Migrations embed.FS

//go:embed templates
var templates embed.FS

//go:embed static/*
var staticFiles embed.FS

func main() {
	flag.Usage = func() {
		log.Info("Usage of Help the stars ⭐")
		log.Info("  https://github.com/ad2ien/help-the-stars/")
		log.Info("")
		log.Info("Flags:")
		flag.PrintDefaults()
	}

	debugFlag := flag.Bool("debug", false, "Display debug logs")
	interval := flag.Int("interval", defaultInterfalHours, "Hours interval between github queries")
	ghTokenFlag := flag.String("gh-token", "", "Github token")
	labels := flag.String("labels",
	 "\"help-wanted\", \"help wanted\",\"junior friendly\",\"good first issue\"", "labels to look for")
	dbFile := flag.String("db-file", "db/help-the-stars.db", "SQLite database file")
	matrixServer := flag.String("matrix-server", "", "Matrix homeserver URL")
	matrixUsername := flag.String("matrix-user", "", "Matrix user")
	matrixPassword := flag.String("matrix-password", "", "Matrix password")
	matrixRoomID := flag.String("matrix-room", "", "Matrix room ID")

	flag.Parse()

	serviceSetting := internal.NewSettingsService(
		&internal.Settings{
			GhToken:        *ghTokenFlag,
			Interval:       *interval,
			Labels:         *labels,
			DBFile:         *dbFile,
			MatrixServer:   *matrixServer,
			MatrixUsername: *matrixUsername,
			MatrixPassword: *matrixPassword,
			MatrixRoomID:   *matrixRoomID,
		})

	serviceSetting.Print()

	matrix := internal.CreateMatrixClient(context.Background(), serviceSetting)

	db := internal.NewConnection(Migrations, serviceSetting)
	defer db.Close()

	controller := internal.CreateController(db.Connection, matrix, serviceSetting)

	log.Info("Start worker...")

	if *debugFlag {
		log.SetLevel(log.DebugLevel)
	}

	log.Debug("Debugging on")

	go controller.Worker()

	server := internal.NewServer(controller, serviceSetting)
	server.Start(&templates, &staticFiles)
}
