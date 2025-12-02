package logger

import (
	"regexp"
	"strings"
)

// MaskAPIKey masks an API key, showing only first 4 and last 4 characters
func MaskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// MaskURL masks sensitive parameters in a URL (key, token, apikey, etc.)
func MaskURL(rawURL string) string {
	// Pattern to match common sensitive query parameters
	patterns := []string{
		`(key=)[^&]+`,
		`(token=)[^&]+`,
		`(apikey=)[^&]+`,
		`(api_key=)[^&]+`,
		`(secret=)[^&]+`,
		`(password=)[^&]+`,
	}

	result := rawURL
	for _, pattern := range patterns {
		re := regexp.MustCompile("(?i)" + pattern)
		result = re.ReplaceAllString(result, "${1}***")
	}

	return result
}

// MaskAuthHeader masks an authorization header value
func MaskAuthHeader(header string) string {
	if header == "" {
		return ""
	}

	// Handle "Bearer xxx" format
	if strings.HasPrefix(header, "Bearer ") {
		return "Bearer ***"
	}

	// Handle "Basic xxx" format
	if strings.HasPrefix(header, "Basic ") {
		return "Basic ***"
	}

	// For other formats, just mask it
	return "***"
}

// MaskString masks a string, showing only first n characters
func MaskString(s string, showFirst int) string {
	if len(s) <= showFirst {
		return "***"
	}
	return s[:showFirst] + "***"
}

// TruncateString truncates a string to maxLen characters, adding "..." if truncated
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}
