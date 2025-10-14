package main

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
)

type Config struct {
	Port           string
	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPassword   string
	RecipientEmail string
	FromEmail      string
	SuccessURL     string
}

func loadConfig() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		SMTPHost:       getEnv("SMTP_HOST", ""),
		SMTPPort:       getEnv("SMTP_PORT", "587"),
		SMTPUser:       getEnv("SMTP_USER", ""),
		SMTPPassword:   getEnv("SMTP_PASSWORD", ""),
		RecipientEmail: getEnv("RECIPIENT_EMAIL", ""),
		FromEmail:      getEnv("FROM_EMAIL", ""),
		SuccessURL:     getEnv("SUCCESS_URL", "/success"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	config := loadConfig()

	// Validate required config
	if config.RecipientEmail == "" {
		log.Fatal("RECIPIENT_EMAIL environment variable is required")
	}
	if config.SMTPHost == "" {
		log.Fatal("SMTP_HOST environment variable is required")
	}
	if config.SMTPUser == "" {
		log.Fatal("SMTP_USER environment variable is required")
	}
	if config.SMTPPassword == "" {
		log.Fatal("SMTP_PASSWORD environment variable is required")
	}
	if config.FromEmail == "" {
		log.Fatal("FROM_EMAIL environment variable is required")
	}

	http.HandleFunc("/submit", handleFormSubmission(config))
	http.HandleFunc("/health", handleHealth)

	log.Printf("Form mailer server starting on port %s", config.Port)
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleFormSubmission(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Build email body from form data
		var emailBody strings.Builder
		emailBody.WriteString("New form submission:\n\n")
		for key, values := range r.PostForm {
			emailBody.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(values, ", ")))
		}

		// Send email asynchronously
		go sendEmail(config, "New Form Submission", emailBody.String())

		// Redirect to success page
		http.Redirect(w, r, config.SuccessURL, http.StatusSeeOther)
	}
}

func sendEmail(config Config, subject, body string) {
	auth := smtp.PlainAuth("", config.SMTPUser, config.SMTPPassword, config.SMTPHost)

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", config.FromEmail, config.RecipientEmail, subject, body))

	addr := fmt.Sprintf("%s:%s", config.SMTPHost, config.SMTPPort)
	err := smtp.SendMail(addr, auth, config.FromEmail, []string{config.RecipientEmail}, msg)
	if err != nil {
		log.Printf("Error sending email: %v", err)
	} else {
		log.Printf("Email sent successfully to %s", config.RecipientEmail)
	}
}
