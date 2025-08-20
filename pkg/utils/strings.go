// Package utils provides common utility functions that can be reused across projects
package utils

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// GenerateRandomString generates a random alphanumeric string of the specified length
func GenerateRandomString(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

// SanitizeString removes unwanted characters from user input
func SanitizeString(input string) string {
	// Trim whitespace and convert to proper case
	sanitized := strings.TrimSpace(input)
	if len(sanitized) == 0 {
		return ""
	}

	// Remove leading/trailing special characters
	sanitized = strings.Trim(sanitized, "!@#$%^&*()_+-=[]{}|;:'\",.<>?/~`")

	return sanitized
}

// TruncateString truncates a string to the specified length, adding "..." if needed
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	if maxLength <= 3 {
		return s[:maxLength]
	}

	return s[:maxLength-3] + "..."
}
