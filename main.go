package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := loadConfig()
	if err := config.Validate(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/submit", handleFormSubmission(config))
	http.HandleFunc("/health", handleHealth)

	log.Printf("Form mailer server starting on port %s", config.Port)
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatal(err)
	}
}
