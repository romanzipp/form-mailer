package main

import (
	"fmt"
	"html"
	"log"
	"net/smtp"
	"strings"
)

func buildEmailBody(formData map[string][]string) string {
	var body strings.Builder
	body.WriteString("<html><body>")
	body.WriteString("<h2>New form submission</h2>")

	for key, values := range formData {
		fieldName := key
		if len(fieldName) > 0 {
			fieldName = strings.ToUpper(string(fieldName[0])) + fieldName[1:]
		}

		value := strings.Join(values, ", ")
		escapedFieldName := html.EscapeString(fieldName)
		escapedValue := html.EscapeString(value)

		if strings.Contains(value, "\n") {
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
	email = strings.ReplaceAll(email, "\r", "")
	email = strings.ReplaceAll(email, "\n", "")
	email = strings.TrimSpace(email)

	if !emailRegex.MatchString(email) {
		return ""
	}
	return email
}

func sendEmail(config Config, subject, body string, formData map[string][]string) {
	auth := smtp.PlainAuth("", config.SMTPUser, config.SMTPPassword, config.SMTPHost)

	fromHeader := config.FromEmail
	if config.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", config.FromName, config.FromEmail)
	}

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n",
		fromHeader, config.RecipientEmail, subject)

	if emailValues, ok := formData["email"]; ok && len(emailValues) > 0 {
		if sanitized := sanitizeEmail(emailValues[0]); sanitized != "" {
			headers += fmt.Sprintf("Reply-To: %s\r\n", sanitized)
		}
	}

	headers += "MIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n"

	msg := []byte(headers + body + "\r\n")
	addr := fmt.Sprintf("%s:%s", config.SMTPHost, config.SMTPPort)
	if err := smtp.SendMail(addr, auth, config.FromEmail, []string{config.RecipientEmail}, msg); err != nil {
		log.Printf("Error sending email: %v", err)
		return
	}
	log.Printf("Email sent successfully to %s", config.RecipientEmail)
}
