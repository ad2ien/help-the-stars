package internal

import (
	"embed"
	"net/http"

	"github.com/charmbracelet/log"
)

type Server struct {
	controller *DataController
	settingsService *SettingsService
}

func NewServer(controller *DataController, settingsService *SettingsService) *Server {
	return &Server{
		controller:        controller,
		settingsService:   settingsService,
	}
}

func (s *Server) Start(templates *embed.FS, staticFiles *embed.FS) {
	log.Info("Start server...")

	webpageHandler := CreateWebpageHandler(s.controller, s.settingsService, templates)

	// Serve static files
	staticHandler := http.FileServer(http.FS(staticFiles))
	http.Handle("/static/", staticHandler)

	http.HandleFunc("/", webpageHandler.HandleWebPage)

	log.Info("Server listening on port 1983")
	err := http.ListenAndServe(":1983", nil)

	if err != nil {
		log.Error("Server starting", "error", err)
	}
}
