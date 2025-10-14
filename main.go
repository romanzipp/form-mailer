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
	FromName       string
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
		FromName:       getEnv("FROM_NAME", ""),
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
		emailBody := buildEmailBody(r.PostForm)

		// Send email asynchronously
		go sendEmail(config, "New Form Submission", emailBody, r.PostForm)

		// Redirect to success page
		http.Redirect(w, r, config.SuccessURL, http.StatusSeeOther)
	}
}

func buildEmailBody(formData map[string][]string) string {
	var body strings.Builder
	body.WriteString("<html><body>")
	body.WriteString("<h2>New form submission</h2>")

	for key, values := range formData {
		// Capitalize first character of field name
		fieldName := key
		if len(fieldName) > 0 {
			fieldName = strings.ToUpper(string(fieldName[0])) + fieldName[1:]
		}

		value := strings.Join(values, ", ")

		// Check if value contains newlines (multiline text)
		if strings.Contains(value, "\n") {
			// Replace newlines with <br> for HTML and add extra spacing
			htmlValue := strings.ReplaceAll(value, "\n", "<br>")
			body.WriteString(fmt.Sprintf("<p><strong>%s:</strong><br>%s</p><br>", fieldName, htmlValue))
		} else {
			body.WriteString(fmt.Sprintf("<p><strong>%s:</strong> %s</p>", fieldName, value))
		}
	}

	body.WriteString("</body></html>")
	return body.String()
}

func sendEmail(config Config, subject, body string, formData map[string][]string) {
	auth := smtp.PlainAuth("", config.SMTPUser, config.SMTPPassword, config.SMTPHost)

	// Format From header with display name if provided
	fromHeader := config.FromEmail
	if config.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", config.FromName, config.FromEmail)
	}

	// Build email headers
	headers := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n", fromHeader, config.RecipientEmail, subject)

	// Add Reply-To if email field exists in form
	if emailValues, ok := formData["email"]; ok && len(emailValues) > 0 && emailValues[0] != "" {
		headers += fmt.Sprintf("Reply-To: %s\r\n", emailValues[0])
	}

	headers += "MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n"

	msg := []byte(headers + body + "\r\n")

	addr := fmt.Sprintf("%s:%s", config.SMTPHost, config.SMTPPort)
	err := smtp.SendMail(addr, auth, config.FromEmail, []string{config.RecipientEmail}, msg)
	if err != nil {
		log.Printf("Error sending email: %v", err)
	} else {
		log.Printf("Email sent successfully to %s", config.RecipientEmail)
	}
}
