package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	SMTPHost        string
	SMTPPort        string
	SMTPUser        string
	SMTPPassword    string
	RecipientEmail  string
	FromEmail       string
	FromName        string
	SuccessURL      string
	ValidationRules map[string][]ValidationRule
}

type ValidationRule struct {
	Name string
	Arg  string
}

func loadConfig() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		SMTPHost:        getEnv("SMTP_HOST", ""),
		SMTPPort:        getEnv("SMTP_PORT", "587"),
		SMTPUser:        getEnv("SMTP_USER", ""),
		SMTPPassword:    getEnv("SMTP_PASSWORD", ""),
		RecipientEmail:  getEnv("RECIPIENT_EMAIL", ""),
		FromEmail:       getEnv("FROM_EMAIL", ""),
		FromName:        getEnv("FROM_NAME", ""),
		SuccessURL:      getEnv("SUCCESS_URL", "/success"),
		ValidationRules: parseValidationRules(getEnv("VALIDATION_RULES", "")),
	}
}

func parseValidationRules(spec string) map[string][]ValidationRule {
	rules := make(map[string][]ValidationRule)
	if spec == "" {
		return rules
	}
	for _, fieldSpec := range strings.Split(spec, ",") {
		fieldSpec = strings.TrimSpace(fieldSpec)
		if fieldSpec == "" {
			continue
		}
		parts := strings.SplitN(fieldSpec, ":", 2)
		if len(parts) != 2 {
			log.Printf("Invalid validation rule (missing ':'): %q", fieldSpec)
			continue
		}
		field := strings.TrimSpace(parts[0])
		if field == "" {
			continue
		}
		for _, ruleSpec := range strings.Split(parts[1], "|") {
			ruleSpec = strings.TrimSpace(ruleSpec)
			if ruleSpec == "" {
				continue
			}
			rule := ValidationRule{}
			if idx := strings.Index(ruleSpec, "."); idx >= 0 {
				rule.Name = ruleSpec[:idx]
				rule.Arg = ruleSpec[idx+1:]
			} else {
				rule.Name = ruleSpec
			}
			rules[field] = append(rules[field], rule)
		}
	}
	return rules
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

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

		// Apply user-defined validation rules
		if err := applyValidationRules(r.PostForm, config.ValidationRules); err != nil {
			log.Printf("Form validation error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
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

var validationEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func applyValidationRules(formData map[string][]string, rules map[string][]ValidationRule) error {
	for field, fieldRules := range rules {
		value := ""
		if values, ok := formData[field]; ok && len(values) > 0 {
			value = strings.TrimSpace(values[0])
		}
		for _, rule := range fieldRules {
			switch rule.Name {
			case "required":
				if value == "" {
					return fmt.Errorf("field '%s' is required", field)
				}
			case "email":
				if value != "" && !validationEmailRegex.MatchString(value) {
					return fmt.Errorf("field '%s' must be a valid email", field)
				}
			case "min":
				n, err := strconv.Atoi(rule.Arg)
				if err != nil {
					return fmt.Errorf("invalid min argument for field '%s'", field)
				}
				if value != "" && len(value) < n {
					return fmt.Errorf("field '%s' must be at least %d characters", field, n)
				}
			case "max":
				n, err := strconv.Atoi(rule.Arg)
				if err != nil {
					return fmt.Errorf("invalid max argument for field '%s'", field)
				}
				if len(value) > n {
					return fmt.Errorf("field '%s' must be at most %d characters", field, n)
				}
			default:
				return fmt.Errorf("unknown validation rule '%s' for field '%s'", rule.Name, field)
			}
		}
	}
	return nil
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
