package internal

import (
	"embed"
	"net/http"
	"time"

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
	server := &http.Server{
		Addr:         ":1983",
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	err := server.ListenAndServe()

	if err != nil {
		log.Error("Server starting", "error", err)
	}
}
