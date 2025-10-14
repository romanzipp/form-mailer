package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
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

		// Validate field lengths
		if err := validateFormData(r.PostForm); err != nil {
			log.Printf("Form validation error: %v", err)
			http.Error(w, "Form validation failed", http.StatusBadRequest)
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

func validateFormData(formData map[string][]string) error {
	const maxFieldLength = 50000 // 50KB per field

	for key, values := range formData {
		for _, value := range values {
			if len(value) > maxFieldLength {
				return fmt.Errorf("field '%s' exceeds maximum length", key)
			}
		}
	}
	return nil
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

		// Escape HTML to prevent injection
		escapedFieldName := html.EscapeString(fieldName)
		escapedValue := html.EscapeString(value)

		// Check if value contains newlines (multiline text)
		if strings.Contains(value, "\n") {
			// Replace newlines with <br> for HTML and add extra spacing
			htmlValue := strings.ReplaceAll(escapedValue, "\n", "<br>")
			body.WriteString(fmt.Sprintf("<p><strong>%s:</strong><br>%s</p><br>", escapedFieldName, htmlValue))
		} else {
			body.WriteString(fmt.Sprintf("<p><strong>%s:</strong> %s</p>", escapedFieldName, escapedValue))
		}
	}

	body.WriteString("</body></html>")
	return body.String()
}

func sanitizeEmail(email string) string {
	// Remove any CR/LF characters to prevent header injection
	email = strings.ReplaceAll(email, "\r", "")
	email = strings.ReplaceAll(email, "\n", "")
	email = strings.TrimSpace(email)

	// Basic email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ""
	}

	return email
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

	// Add Reply-To if email field exists in form and is valid
	if emailValues, ok := formData["email"]; ok && len(emailValues) > 0 {
		sanitizedEmail := sanitizeEmail(emailValues[0])
		if sanitizedEmail != "" {
			headers += fmt.Sprintf("Reply-To: %s\r\n", sanitizedEmail)
		}
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
