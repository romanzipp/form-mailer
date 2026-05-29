package main

import (
	"errors"
	"os"
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
	BackURL         string
	ValidationRules map[string][]ValidationRule
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
		BackURL:         getEnv("BACK_URL", ""),
		ValidationRules: parseValidationRules(getEnv("VALIDATION_RULES", "")),
	}
}

func (c Config) Validate() error {
	switch {
	case c.RecipientEmail == "":
		return errors.New("RECIPIENT_EMAIL environment variable is required")
	case c.SMTPHost == "":
		return errors.New("SMTP_HOST environment variable is required")
	case c.SMTPUser == "":
		return errors.New("SMTP_USER environment variable is required")
	case c.SMTPPassword == "":
		return errors.New("SMTP_PASSWORD environment variable is required")
	case c.FromEmail == "":
		return errors.New("FROM_EMAIL environment variable is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
