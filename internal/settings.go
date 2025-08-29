package internal

import (
	"os"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

var once sync.Once

func GetSettings() {
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Failed to load .env file")
		}
	})
}

func GetSetting(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Warn("Missing environment variable: " + key)
		return ""
	}
	return value
}
