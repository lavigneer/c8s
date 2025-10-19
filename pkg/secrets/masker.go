/*
Copyright 2025 C8S Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package secrets

import (
	"bytes"
	"regexp"
	"strings"
)

const redactedMarker = "***REDACTED***"

// MaskSecrets replaces all occurrences of secret values in logs with a redacted marker.
// It performs case-insensitive matching to catch secret values even if they're transformed.
func MaskSecrets(logs []byte, secrets map[string]string) []byte {
	if len(logs) == 0 || len(secrets) == 0 {
		return logs
	}

	masked := logs

	// Sort secrets by length (longest first) to avoid partial replacements
	// For example, if we have secrets "abc123" and "abc", we want to replace "abc123" first
	secretValues := make([]string, 0, len(secrets))
	for _, value := range secrets {
		if value != "" {
			secretValues = append(secretValues, value)
		}
	}

	// Sort by length descending
	for i := 0; i < len(secretValues); i++ {
		for j := i + 1; j < len(secretValues); j++ {
			if len(secretValues[j]) > len(secretValues[i]) {
				secretValues[i], secretValues[j] = secretValues[j], secretValues[i]
			}
		}
	}

	// Replace each secret value (case-insensitive)
	for _, secretValue := range secretValues {
		// Escape special regex characters in the secret value
		escapedSecret := regexp.QuoteMeta(secretValue)

		// Create case-insensitive regex
		pattern := "(?i)" + escapedSecret
		re := regexp.MustCompile(pattern)

		masked = re.ReplaceAll(masked, []byte(redactedMarker))
	}

	return masked
}

// MaskSecretsString is a convenience function that works with strings
func MaskSecretsString(logs string, secrets map[string]string) string {
	return string(MaskSecrets([]byte(logs), secrets))
}

// ExtractSecretValues extracts the actual secret values from a map of secret names to values.
// This is useful when you have a map where keys are secret identifiers and values are the actual secrets.
func ExtractSecretValues(secretData map[string][]byte) map[string]string {
	result := make(map[string]string, len(secretData))
	for key, value := range secretData {
		result[key] = string(value)
	}
	return result
}

// HasRedactedContent checks if the logs contain any redacted markers
func HasRedactedContent(logs []byte) bool {
	return bytes.Contains(logs, []byte(redactedMarker))
}

// CountRedactions returns the number of times secrets were redacted in the logs
func CountRedactions(logs []byte) int {
	return bytes.Count(logs, []byte(redactedMarker))
}

// SanitizeForDisplay prepares logs for display by ensuring all secrets are masked
// and optionally truncating if too long
func SanitizeForDisplay(logs []byte, secrets map[string]string, maxLength int) []byte {
	masked := MaskSecrets(logs, secrets)

	if maxLength > 0 && len(masked) > maxLength {
		// Truncate with a message
		truncated := masked[:maxLength]
		truncated = append(truncated, []byte("\n...[truncated]...")...)
		return truncated
	}

	return masked
}

// IsLikelySecretValue performs heuristic checks to determine if a string looks like a secret.
// This is useful for additional protection when you don't have an explicit list of secrets.
func IsLikelySecretValue(value string) bool {
	value = strings.TrimSpace(value)

	// Too short to be a meaningful secret
	if len(value) < 8 {
		return false
	}

	// Check for common secret patterns
	patterns := []string{
		`^[A-Za-z0-9+/]{40,}={0,2}$`,            // Base64-like
		`^[a-f0-9]{32,}$`,                       // Hex tokens (MD5, SHA-like)
		`^[A-Z0-9_]{20,}$`,                      // API keys
		`^sk-[a-zA-Z0-9]{32,}$`,                 // OpenAI-style keys
		`^gh[ps]_[a-zA-Z0-9]{36,}$`,             // GitHub tokens
		`^xox[baprs]-[a-zA-Z0-9-]{10,}$`,        // Slack tokens
		`^AKIA[0-9A-Z]{16}$`,                    // AWS Access Key ID
		`^[a-zA-Z0-9/+=]{40}$`,                  // AWS Secret Access Key
		`^ya29\.[a-zA-Z0-9_-]{50,}$`,            // Google OAuth tokens
		`^eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+`, // JWT tokens
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, value); matched {
			return true
		}
	}

	return false
}
