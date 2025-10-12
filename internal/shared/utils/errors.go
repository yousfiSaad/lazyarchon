package utils

import (
	"fmt"
	"strings"
)

// FormatUserFriendlyError converts technical errors to user-friendly messages
func FormatUserFriendlyError(err string) string {
	if err == "" {
		return ""
	}

	// Convert common error patterns to user-friendly messages
	lowercaseErr := strings.ToLower(err)

	if strings.Contains(lowercaseErr, "connection refused") ||
		strings.Contains(lowercaseErr, "no such host") ||
		strings.Contains(lowercaseErr, "network is unreachable") {
		return "Unable to connect to server. Please check your network connection and server settings."
	}

	if strings.Contains(lowercaseErr, "unauthorized") ||
		strings.Contains(lowercaseErr, "invalid api key") ||
		strings.Contains(lowercaseErr, "authentication failed") {
		return "Authentication failed. Please check your API key in the configuration."
	}

	if strings.Contains(lowercaseErr, "timeout") ||
		strings.Contains(lowercaseErr, "deadline exceeded") {
		return "Request timed out. The server may be overloaded or your connection is slow."
	}

	if strings.Contains(lowercaseErr, "not found") {
		return "The requested resource was not found on the server."
	}

	if strings.Contains(lowercaseErr, "bad request") ||
		strings.Contains(lowercaseErr, "invalid request") {
		return "Invalid request. Please check your input and try again."
	}

	// If no specific pattern matches, return a generic message with the original error
	return fmt.Sprintf("An error occurred: %s", err)
}
