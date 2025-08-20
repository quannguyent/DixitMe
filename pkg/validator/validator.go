// Package validator provides common validation functions for user input
package validator

import (
	"regexp"
	"strings"
)

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
