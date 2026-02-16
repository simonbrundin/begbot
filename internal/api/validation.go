package api

import (
	"strconv"
	"strings"
)

func ValidateRequired(value interface{}, fieldName string) []ValidationError {
	var errors []ValidationError
	if value == nil {
		errors = append(errors, ValidationError{Field: fieldName, Message: "required"})
		return errors
	}
	if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
		errors = append(errors, ValidationError{Field: fieldName, Message: "required"})
	}
	return errors
}

func ValidateMinLength(value string, fieldName string, min int) []ValidationError {
	var errors []ValidationError
	if len(strings.TrimSpace(value)) < min {
		errors = append(errors, ValidationError{Field: fieldName, Message: "must be at least " + strconv.Itoa(min) + " characters"})
	}
	return errors
}

func ValidateNonNegative(value int64, fieldName string) []ValidationError {
	var errors []ValidationError
	if value < 0 {
		errors = append(errors, ValidationError{Field: fieldName, Message: "must be non-negative"})
	}
	return errors
}

func CombineErrors(errorLists ...[]ValidationError) []ValidationError {
	var combined []ValidationError
	for _, list := range errorLists {
		combined = append(combined, list...)
	}
	return combined
}
