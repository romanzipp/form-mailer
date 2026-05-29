package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

func logSubmission(formData map[string][]string) {
	keys := make([]string, 0, len(formData))
	for k := range formData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		total := 0
		for _, v := range formData[k] {
			total += len(v)
		}
		parts = append(parts, fmt.Sprintf("%s=%d", k, total))
	}
	log.Printf("Form submission received: %s", strings.Join(parts, " "))
}
