package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

const maxFieldLength = 50000 // 50KB per field

type ValidationRule struct {
	Name string
	Arg  string
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

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

func validateFieldLengths(formData map[string][]string) error {
	for key, values := range formData {
		for _, value := range values {
			if len(value) > maxFieldLength {
				return fmt.Errorf("field '%s' exceeds maximum length", key)
			}
		}
	}
	return nil
}

func applyValidationRules(formData map[string][]string, rules map[string][]ValidationRule) error {
	for field, fieldRules := range rules {
		value := ""
		if values, ok := formData[field]; ok && len(values) > 0 {
			value = strings.TrimSpace(values[0])
		}
		for _, rule := range fieldRules {
			if err := checkRule(field, value, rule); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkRule(field, value string, rule ValidationRule) error {
	switch rule.Name {
	case "required":
		if value == "" {
			return fmt.Errorf("field '%s' is required", field)
		}
	case "email":
		if value != "" && !emailRegex.MatchString(value) {
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
	return nil
}
