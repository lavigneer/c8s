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

package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/org/c8s/pkg/secrets"
)

func TestMaskSecrets_SingleSecret(t *testing.T) {
	logs := []byte("Starting deployment with API key: mysecretkey123")
	secretValues := map[string]string{
		"api_key": "mysecretkey123",
	}

	masked := secrets.MaskSecrets(logs, secretValues)

	expected := "Starting deployment with API key: ***REDACTED***"
	assert.Equal(t, expected, string(masked))
	assert.True(t, secrets.HasRedactedContent(masked))
	assert.Equal(t, 1, secrets.CountRedactions(masked))
}

func TestMaskSecrets_MultipleSecrets(t *testing.T) {
	logs := []byte("API key: secretkey123, Password: pass456, Token: token789")
	secretValues := map[string]string{
		"api_key":  "secretkey123",
		"password": "pass456",
		"token":    "token789",
	}

	masked := secrets.MaskSecrets(logs, secretValues)

	expected := "API key: ***REDACTED***, Password: ***REDACTED***, Token: ***REDACTED***"
	assert.Equal(t, expected, string(masked))
	assert.True(t, secrets.HasRedactedContent(masked))
	assert.Equal(t, 3, secrets.CountRedactions(masked))
}

func TestMaskSecrets_PartialSecretValue(t *testing.T) {
	logs := []byte("My password is: secret123 and it's very secure")
	secretValues := map[string]string{
		"password": "secret123",
	}

	masked := secrets.MaskSecrets(logs, secretValues)

	expected := "My password is: ***REDACTED*** and it's very secure"
	assert.Equal(t, expected, string(masked))
}

func TestMaskSecrets_CaseInsensitive(t *testing.T) {
	logs := []byte("Secret: MySecretKey123 and mysecretkey123 and MYSECRETKEY123")
	secretValues := map[string]string{
		"secret": "MySecretKey123",
	}

	masked := secrets.MaskSecrets(logs, secretValues)

	// All variations should be redacted (case-insensitive)
	expected := "Secret: ***REDACTED*** and ***REDACTED*** and ***REDACTED***"
	assert.Equal(t, expected, string(masked))
	assert.Equal(t, 3, secrets.CountRedactions(masked))
}

func TestMaskSecrets_SpecialCharacters(t *testing.T) {
	logs := []byte("Token: secret$123@key! is used")
	secretValues := map[string]string{
		"token": "secret$123@key!",
	}

	masked := secrets.MaskSecrets(logs, secretValues)

	expected := "Token: ***REDACTED*** is used"
	assert.Equal(t, expected, string(masked))
}

func TestMaskSecrets_EmptyLogs(t *testing.T) {
	logs := []byte("")
	secretValues := map[string]string{
		"secret": "value",
	}

	masked := secrets.MaskSecrets(logs, secretValues)

	assert.Equal(t, []byte(""), masked)
	assert.False(t, secrets.HasRedactedContent(masked))
}

func TestMaskSecrets_NoSecrets(t *testing.T) {
	logs := []byte("This is a normal log line without secrets")
	secretValues := map[string]string{}

	masked := secrets.MaskSecrets(logs, secretValues)

	assert.Equal(t, logs, masked)
	assert.False(t, secrets.HasRedactedContent(masked))
}

func TestMaskSecrets_SecretNotInLogs(t *testing.T) {
	logs := []byte("This is a normal log line")
	secretValues := map[string]string{
		"secret": "mysecret",
	}

	masked := secrets.MaskSecrets(logs, secretValues)

	assert.Equal(t, logs, masked)
	assert.False(t, secrets.HasRedactedContent(masked))
}

func TestMaskSecrets_MultipleOccurrences(t *testing.T) {
	logs := []byte("Secret: key123, using key123 again, and key123 once more")
	secretValues := map[string]string{
		"secret": "key123",
	}

	masked := secrets.MaskSecrets(logs, secretValues)

	expected := "Secret: ***REDACTED***, using ***REDACTED*** again, and ***REDACTED*** once more"
	assert.Equal(t, expected, string(masked))
	assert.Equal(t, 3, secrets.CountRedactions(masked))
}

func TestMaskSecrets_LongerSecretsFirst(t *testing.T) {
	// Test that longer secrets are masked before shorter ones
	// This prevents "abc123" being partially redacted before "abc123456" can be matched
	logs := []byte("Long secret: abc123456, Short secret: abc123")
	secretValues := map[string]string{
		"short": "abc123",
		"long":  "abc123456",
	}

	masked := secrets.MaskSecrets(logs, secretValues)

	// Both should be fully redacted
	expected := "Long secret: ***REDACTED***, Short secret: ***REDACTED***"
	assert.Equal(t, expected, string(masked))
	assert.Equal(t, 2, secrets.CountRedactions(masked))
}

func TestMaskSecrets_MultilineLog(t *testing.T) {
	logs := []byte(`Starting process
API Key: secret123
Connecting to server
Password: pass456
Done`)
	secretValues := map[string]string{
		"api_key":  "secret123",
		"password": "pass456",
	}

	masked := secrets.MaskSecrets(logs, secretValues)

	expected := `Starting process
API Key: ***REDACTED***
Connecting to server
Password: ***REDACTED***
Done`
	assert.Equal(t, expected, string(masked))
	assert.Equal(t, 2, secrets.CountRedactions(masked))
}

func TestMaskSecretsString(t *testing.T) {
	logs := "API Key: secretkey123"
	secretValues := map[string]string{
		"api_key": "secretkey123",
	}

	masked := secrets.MaskSecretsString(logs, secretValues)

	expected := "API Key: ***REDACTED***"
	assert.Equal(t, expected, masked)
}

func TestExtractSecretValues(t *testing.T) {
	secretData := map[string][]byte{
		"api_key":  []byte("secret123"),
		"password": []byte("pass456"),
	}

	result := secrets.ExtractSecretValues(secretData)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, "secret123", result["api_key"])
	assert.Equal(t, "pass456", result["password"])
}

func TestSanitizeForDisplay(t *testing.T) {
	logs := []byte("API Key: secret123 and more text")
	secretValues := map[string]string{
		"api_key": "secret123",
	}

	// Test without truncation
	masked := secrets.SanitizeForDisplay(logs, secretValues, 0)
	assert.Contains(t, string(masked), "***REDACTED***")
	assert.NotContains(t, string(masked), "secret123")

	// Test with truncation
	masked = secrets.SanitizeForDisplay(logs, secretValues, 20)
	assert.True(t, len(masked) <= 40) // 20 + some overhead for truncation message
	assert.Contains(t, string(masked), "truncated")
}

func TestIsLikelySecretValue_Base64(t *testing.T) {
	assert.True(t, secrets.IsLikelySecretValue("dGhpcyBpcyBhIHNlY3JldCB0aGF0IGlzIGJhc2U2NCBlbmNvZGVk"))
	assert.False(t, secrets.IsLikelySecretValue("short"))
}

func TestIsLikelySecretValue_HexToken(t *testing.T) {
	assert.True(t, secrets.IsLikelySecretValue("a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6"))
	assert.False(t, secrets.IsLikelySecretValue("not-hex-just-text"))
}

func TestIsLikelySecretValue_APIKey(t *testing.T) {
	assert.True(t, secrets.IsLikelySecretValue("AKIAIOSFODNN7EXAMPLE"))
	assert.False(t, secrets.IsLikelySecretValue("normaltext"))
}

func TestIsLikelySecretValue_GitHubToken(t *testing.T) {
	assert.True(t, secrets.IsLikelySecretValue("ghp_1234567890abcdefghijklmnopqrstuvwxyz"))
	assert.True(t, secrets.IsLikelySecretValue("ghs_1234567890abcdefghijklmnopqrstuvwxyz"))
}

func TestIsLikelySecretValue_JWT(t *testing.T) {
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	assert.True(t, secrets.IsLikelySecretValue(jwt))
}

func TestIsLikelySecretValue_TooShort(t *testing.T) {
	assert.False(t, secrets.IsLikelySecretValue("short"))
	assert.False(t, secrets.IsLikelySecretValue("1234567")) // Less than 8 characters
}

func TestMaskSecrets_EmptySecretValue(t *testing.T) {
	logs := []byte("API Key: secret123")
	secretValues := map[string]string{
		"empty":   "",
		"api_key": "secret123",
	}

	masked := secrets.MaskSecrets(logs, secretValues)

	expected := "API Key: ***REDACTED***"
	assert.Equal(t, expected, string(masked))
	assert.Equal(t, 1, secrets.CountRedactions(masked))
}

func TestMaskSecrets_NewlinesInSecret(t *testing.T) {
	// Test that secrets containing newlines are handled properly
	logs := []byte("Certificate:\n-----BEGIN-----\ncert123\n-----END-----\nDone")
	secretValues := map[string]string{
		"cert": "-----BEGIN-----\ncert123\n-----END-----",
	}

	masked := secrets.MaskSecrets(logs, secretValues)

	assert.Contains(t, string(masked), "***REDACTED***")
	assert.NotContains(t, string(masked), "cert123")
}
