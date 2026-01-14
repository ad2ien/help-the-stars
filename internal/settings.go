package internal

import (
	"regexp"

	"github.com/charmbracelet/log"
)

var settings *Settings

type Settings struct {
	GhToken        string
	Labels         string
	MatrixRoomID   string
	MatrixUsername string
	MatrixPassword string
	MatrixServer   string
	DBFile         string
}

func GetSettings() *Settings {
	return settings
}

func SetSettings(localSettings *Settings) {
	localSettings.checkSettings()
	settings = localSettings
}

func (s *Settings) IsMatrixConfigured() bool {
	return s.MatrixRoomID != "" &&
		s.MatrixUsername != "" &&
		s.MatrixPassword != "" &&
		s.MatrixServer != ""
}

func (s *Settings) checkSettings() {
	if s.GhToken == "" {
		log.Fatal("Missing Github token")
	}
	if s.DBFile == "" {
		log.Fatal("Missing database file path")
	}
	if !s.IsMatrixConfigured() {
		log.Warn("Matrix notif not fully configured")
	}
	if s.Labels == "" {
		log.Fatal("Missing labels")
	}
	if s.GetLabelSlice() == nil {
		log.Fatal("Invalid labels, format should be \"label1\", \"label2\", ...")
	}
}

func (s *Settings) GetLabelSlice() []string {
	// Regex to extract labels inside double quotes
	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindAllStringSubmatch(s.Labels, -1)

	var labels []string
	for _, match := range matches {
		labels = append(labels, match[1])
	}
	return labels
}
