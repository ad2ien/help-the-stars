package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var once sync.Once

func GetSettings() {
	once.Do(func() {
		fmt.Println("Only see this once")
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Failed to load .env file")
		}
	})
}

func GetSetting(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatal("Missing environment variable: " + key)
	}
	return value
}
