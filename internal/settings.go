package internal

import (
	"regexp"
	"strconv"

	"github.com/charmbracelet/log"
)

const hoursInSeconds = 3600

// SettingsService provides access to app settings.
type SettingsService struct {
	settings *Settings
}

// NewSettingsService creates a new SettingsService instance.
func NewSettingsService(settings *Settings) *SettingsService {
	settings.checkSettings()

	return &SettingsService{
		settings: settings,
	}
}

// Settings of the app.
type Settings struct {
	GhToken string
	// Hours interval between Github queries
	Interval int
	DBFile   string
	// labels to look for
	// ex: "\"help-wanted\", \"help wanted\",\"junior friendly\",\"good first issue\""
	Labels string
	// optional
	MatrixRoomID   string
	MatrixUsername string
	MatrixPassword string
	MatrixServer   string

	maxAge string
}

// GetSettings : Get all the settings.
func (ss *SettingsService) GetSettings() *Settings {
	return ss.settings
}

// Print settings.
func (ss *SettingsService) Print() {
	log.Info("Settings:")
	log.Info("  Github token: **")
	log.Infof("  Interval: %d hours", ss.GetSettings().Interval)
	log.Infof("  Database file: %s", ss.GetSettings().DBFile)
	log.Infof("  Labels: %s", ss.GetSettings().Labels)
	log.Infof("  Matrix room ID: %s", ss.GetSettings().MatrixRoomID)
	log.Infof("  Matrix username: %s", ss.GetSettings().MatrixUsername)
	log.Info("  Matrix password: ***")
	log.Infof("  Matrix server: %s", ss.GetSettings().MatrixServer)
}

// IsMatrixConfigured checks if matrix configuration is complete.
func (s *Settings) IsMatrixConfigured() bool {
	return s.MatrixRoomID != "" &&
		s.MatrixUsername != "" &&
		s.MatrixPassword != "" &&
		s.MatrixServer != ""
}

// GetLabelSlice gets label slice from settings.
func (s *Settings) GetLabelSlice() []string {
	// Regex to extract labels inside double quotes
	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindAllStringSubmatch(s.Labels, -1)

	var labels = make([]string, len(matches))
	for _, match := range matches {
		labels = append(labels, match[1])
	}

	return labels
}

// GetMaxAge gets the max time before which data is not fresh anymore.
func (ss *SettingsService) GetMaxAge() string {
	s := ss.GetSettings()
	if s.maxAge == "" {
		s.maxAge = strconv.Itoa(hoursInSeconds * s.Interval)
	}

	return s.maxAge
}

func (s *Settings) checkSettings() {
	if s.GhToken == "" {
		log.Fatal("Missing Github token")
	}

	if s.Interval < 1 || s.Interval > 100 {
		log.Fatal("Invalid interval : should be between 1 and 100 hours")
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
