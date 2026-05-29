package main

import (
	"log"
	"net/http"
)

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

		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			renderServerError(w, config, "Could not parse form data.")
			return
		}

		if err := validateFieldLengths(r.PostForm); err != nil {
			log.Printf("Form validation error: %v", err)
			renderValidationError(w, config, r.PostForm, err.Error())
			return
		}

		if err := applyValidationRules(r.PostForm, config.ValidationRules); err != nil {
			log.Printf("Form validation error: %v", err)
			renderValidationError(w, config, r.PostForm, err.Error())
			return
		}

		logSubmission(r.PostForm)

		emailBody := buildEmailBody(r.PostForm)
		go sendEmail(config, "New Form Submission", emailBody, r.PostForm)

		http.Redirect(w, r, config.SuccessURL, http.StatusSeeOther)
	}
}
