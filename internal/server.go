package internal

import (
	"embed"
	"net/http"

	"github.com/charmbracelet/log"
)

type Server struct {
	controller *DataController
}

func NewServer(controller *DataController) *Server {
	return &Server{
		controller: controller,
	}
}

func (s *Server) Start(templates *embed.FS, staticFiles *embed.FS) {
	log.Info("Start server...")

	webpageHandler := CreateWebpageHandler(s.controller, templates)

	// Serve static files
	staticHandler := http.FileServer(http.FS(staticFiles))
	http.Handle("/static/", staticHandler)

	http.HandleFunc("/", webpageHandler.HandleWebPage)

	log.Info("Server listening on port 1983")
	err := http.ListenAndServe(":1983", nil)

	if err != nil {
		log.Error("Server error", err)
	}
}
