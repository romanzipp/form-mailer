package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
)

//go:embed templates/*.html
var templatesFS embed.FS

var templates = template.Must(template.ParseFS(templatesFS, "templates/*.html"))

type formField struct {
	Name      string
	Label     string
	Value     string
	Multiline bool
}

type validationErrorPage struct {
	Error   string
	Fields  []formField
	BackURL string
	Action  string
}

type serverErrorPage struct {
	Error   string
	BackURL string
}

func renderValidationError(w http.ResponseWriter, config Config, formData map[string][]string, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	data := validationErrorPage{
		Error:   errMsg,
		Fields:  buildFormFields(formData),
		BackURL: config.BackURL,
		Action:  "/submit",
	}
	if err := templates.ExecuteTemplate(w, "validation_error", data); err != nil {
		log.Printf("Error rendering validation page: %v", err)
	}
}

func renderServerError(w http.ResponseWriter, config Config, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	data := serverErrorPage{Error: errMsg, BackURL: config.BackURL}
	if err := templates.ExecuteTemplate(w, "server_error", data); err != nil {
		log.Printf("Error rendering server error page: %v", err)
	}
}

func buildFormFields(formData map[string][]string) []formField {
	keys := make([]string, 0, len(formData))
	for k := range formData {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fields := make([]formField, 0, len(keys))
	for _, k := range keys {
		value := strings.Join(formData[k], ", ")
		label := k
		if len(label) > 0 {
			label = strings.ToUpper(string(label[0])) + label[1:]
		}
		fields = append(fields, formField{
			Name:      k,
			Label:     label,
			Value:     value,
			Multiline: strings.Contains(value, "\n"),
		})
	}
	return fields
}
