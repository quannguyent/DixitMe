// Package utils provides common utility functions for the DixitMe application
package utils

import (
	"crypto/rand"
	"math/big"
	"regexp"
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

// Validation regular expressions
var (
	// EmailRegex validates email format
	EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// UsernameRegex validates username format (alphanumeric + underscore, 3-20 chars)
	UsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

	// RoomCodeRegex validates room code format (6 alphanumeric characters)
	RoomCodeRegex = regexp.MustCompile(`^[A-Z0-9]{6}$`)
)

// ValidateEmail checks if an email address is valid
func ValidateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) == 0 || len(email) > 254 {
		return false
	}
	return EmailRegex.MatchString(email)
}

// ValidateUsername checks if a username is valid
func ValidateUsername(username string) bool {
	username = strings.TrimSpace(username)
	return UsernameRegex.MatchString(username)
}

// ValidatePassword checks if a password meets requirements
func ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "password must be at least 8 characters long"
	}
	if len(password) > 128 {
		return false, "password must be less than 128 characters long"
	}

	// Check for at least one letter and one number
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasLetter || !hasNumber {
		return false, "password must contain at least one letter and one number"
	}

	return true, ""
}

// ValidateRoomCode checks if a room code is valid
func ValidateRoomCode(roomCode string) bool {
	roomCode = strings.TrimSpace(strings.ToUpper(roomCode))
	return RoomCodeRegex.MatchString(roomCode)
}

// ValidateDisplayName checks if a display name is valid
func ValidateDisplayName(name string) bool {
	name = strings.TrimSpace(name)
	return len(name) >= 1 && len(name) <= 50
}
