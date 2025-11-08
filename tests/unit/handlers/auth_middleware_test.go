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

package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authpkg "github.com/org/c8s/cmd/api-server/auth"
	handlerspkg "github.com/org/c8s/cmd/api-server/handlers"
)

// testSetup resets validator state before each test to prevent cross-test contamination
func testSetup(t *testing.T) {
	handlerspkg.ResetAuthValidators()
	t.Cleanup(func() {
		handlerspkg.ResetAuthValidators()
	})
}

// createTestToken creates a valid JWT token for testing
func createTestToken(secret string, issuer string, audience string, subject string) string {
	now := time.Now()
	claims := authpkg.Claims{
		Subject:   subject,
		Name:      "Test User",
		Email:     "test@example.com",
		Namespace: "test-ns",
		Roles:     []string{"viewer"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

// TestAuthMiddlewareWithValidToken tests successful authentication with valid JWT
func TestAuthMiddlewareWithValidToken(t *testing.T) {
	testSetup(t)
	secret := "test-secret-key-should-be-longer-32-bytes"
	config := &authpkg.Config{
		Mode:             "jwt",
		Algorithm:        "HS256",
		Issuer:           "c8s-auth",
		Audience:         "c8s-api",
		Secret:           secret,
		VerifyExpiry:     true,
		VerifySignature:  true,
		ExpiryTolerance:  0,
		DefaultNamespace: "default",
		DefaultRoles:     []string{"user"},
	}

	// Initialize validator
	err := handlerspkg.InitAuthValidator(config)
	require.NoError(t, err)

	// Create a valid token
	tokenString := createTestToken(secret, "c8s-auth", "c8s-api", "user-123")

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := handlerspkg.GetUserFromContext(r.Context())
		require.True(t, ok, "user should be in context")
		assert.Equal(t, "user-123", user.ID)
		assert.Equal(t, "Test User", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "test-ns", user.Namespace)
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware
	handler := handlerspkg.AuthMiddleware(testHandler)

	// Test with Authorization header
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAuthMiddlewareWithAuthCookie tests authentication via auth_token cookie
func TestAuthMiddlewareWithAuthCookie(t *testing.T) {
	testSetup(t)
	secret := "test-secret-key-should-be-longer-32-bytes"
	config := &authpkg.Config{
		Mode:             "jwt",
		Algorithm:        "HS256",
		Issuer:           "c8s-auth",
		Audience:         "c8s-api",
		Secret:           secret,
		VerifyExpiry:     true,
		VerifySignature:  true,
		ExpiryTolerance:  0,
		DefaultNamespace: "default",
		DefaultRoles:     []string{"user"},
	}

	err := handlerspkg.InitAuthValidator(config)
	require.NoError(t, err)

	tokenString := createTestToken(secret, "c8s-auth", "c8s-api", "user-456")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := handlerspkg.GetUserFromContext(r.Context())
		require.True(t, ok)
		assert.Equal(t, "user-456", user.ID)
		w.WriteHeader(http.StatusOK)
	})

	handler := handlerspkg.AuthMiddleware(testHandler)

	// Test with cookie
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: tokenString,
	})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAuthMiddlewareWithInvalidToken tests rejection of invalid tokens
func TestAuthMiddlewareWithInvalidToken(t *testing.T) {
	testSetup(t)
	secret := "test-secret-key-should-be-longer-32-bytes"
	config := &authpkg.Config{
		Mode:             "jwt",
		Algorithm:        "HS256",
		Issuer:           "c8s-auth",
		Audience:         "c8s-api",
		Secret:           secret,
		VerifyExpiry:     true,
		VerifySignature:  true,
		ExpiryTolerance:  0,
		DefaultNamespace: "default",
		DefaultRoles:     []string{"user"},
	}

	err := handlerspkg.InitAuthValidator(config)
	require.NoError(t, err)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := handlerspkg.AuthMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAuthMiddlewareNoTokenAPIRequest tests API request without token
func TestAuthMiddlewareNoTokenAPIRequest(t *testing.T) {
	testSetup(t)
	secret := "test-secret-key-should-be-longer-32-bytes"
	config := &authpkg.Config{
		Mode:             "jwt",
		Algorithm:        "HS256",
		Issuer:           "c8s-auth",
		Audience:         "c8s-api",
		Secret:           secret,
		VerifyExpiry:     true,
		VerifySignature:  true,
		ExpiryTolerance:  0,
		DefaultNamespace: "default",
		DefaultRoles:     []string{"user"},
	}

	err := handlerspkg.InitAuthValidator(config)
	require.NoError(t, err)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := handlerspkg.AuthMiddleware(testHandler)

	// API request (Accept: application/json) without token
	req := httptest.NewRequest("GET", "/api/test", http.NoBody)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAuthMiddlewareNoTokenHTMLRequest tests HTML request without token redirects to login
func TestAuthMiddlewareNoTokenHTMLRequest(t *testing.T) {
	testSetup(t)
	secret := "test-secret-key-should-be-longer-32-bytes"
	config := &authpkg.Config{
		Mode:             "jwt",
		Algorithm:        "HS256",
		Issuer:           "c8s-auth",
		Audience:         "c8s-api",
		Secret:           secret,
		VerifyExpiry:     true,
		VerifySignature:  true,
		ExpiryTolerance:  0,
		DefaultNamespace: "default",
		DefaultRoles:     []string{"user"},
	}

	err := handlerspkg.InitAuthValidator(config)
	require.NoError(t, err)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := handlerspkg.AuthMiddleware(testHandler)

	// HTML request (Accept: text/html) without token
	req := httptest.NewRequest("GET", "/dashboard", http.NoBody)
	req.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))
}

// TestOptionalAuthMiddlewareWithValidToken tests optional auth with valid token
func TestOptionalAuthMiddlewareWithValidToken(t *testing.T) {
	testSetup(t)
	secret := "test-secret-key-should-be-longer-32-bytes"
	config := &authpkg.Config{
		Mode:             "jwt",
		Algorithm:        "HS256",
		Issuer:           "c8s-auth",
		Audience:         "c8s-api",
		Secret:           secret,
		VerifyExpiry:     true,
		VerifySignature:  true,
		ExpiryTolerance:  0,
		DefaultNamespace: "default",
		DefaultRoles:     []string{"user"},
	}

	err := handlerspkg.InitAuthValidator(config)
	require.NoError(t, err)

	tokenString := createTestToken(secret, "c8s-auth", "c8s-api", "user-789")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := handlerspkg.GetUserFromContext(r.Context())
		require.True(t, ok, "user should be in context")
		assert.Equal(t, "user-789", user.ID)
		w.WriteHeader(http.StatusOK)
	})

	handler := handlerspkg.OptionalAuthMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestOptionalAuthMiddlewareWithoutToken tests optional auth works without token
func TestOptionalAuthMiddlewareWithoutToken(t *testing.T) {
	testSetup(t)
	secret := "test-secret-key-should-be-longer-32-bytes"
	config := &authpkg.Config{
		Mode:             "jwt",
		Algorithm:        "HS256",
		Issuer:           "c8s-auth",
		Audience:         "c8s-api",
		Secret:           secret,
		VerifyExpiry:     true,
		VerifySignature:  true,
		ExpiryTolerance:  0,
		DefaultNamespace: "default",
		DefaultRoles:     []string{"user"},
	}

	err := handlerspkg.InitAuthValidator(config)
	require.NoError(t, err)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// User should NOT be in context
		user, ok := handlerspkg.GetUserFromContext(r.Context())
		assert.False(t, ok, "user should not be in context")
		assert.Nil(t, user)
		w.WriteHeader(http.StatusOK)
	})

	handler := handlerspkg.OptionalAuthMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	// No authorization header or cookie
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestOptionalAuthMiddlewareWithInvalidToken continues without auth on invalid token
func TestOptionalAuthMiddlewareWithInvalidToken(t *testing.T) {
	testSetup(t)
	secret := "test-secret-key-should-be-longer-32-bytes"
	config := &authpkg.Config{
		Mode:             "jwt",
		Algorithm:        "HS256",
		Issuer:           "c8s-auth",
		Audience:         "c8s-api",
		Secret:           secret,
		VerifyExpiry:     true,
		VerifySignature:  true,
		ExpiryTolerance:  0,
		DefaultNamespace: "default",
		DefaultRoles:     []string{"user"},
	}

	err := handlerspkg.InitAuthValidator(config)
	require.NoError(t, err)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// User should NOT be in context because token was invalid
		user, ok := handlerspkg.GetUserFromContext(r.Context())
		assert.False(t, ok, "user should not be in context")
		assert.Nil(t, user)
		w.WriteHeader(http.StatusOK)
	})

	handler := handlerspkg.OptionalAuthMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestNoOpValidator tests development mode validator
func TestNoOpValidator(t *testing.T) {
	testSetup(t)
	// Use NoOp validator for development
	handlerspkg.UseNoOpValidator()

	// Note: Reset would require reinitializing in real test suite
	// This is just to verify the function exists and doesn't panic

	secret := "test-secret"
	config := &authpkg.Config{
		Mode:      "jwt",
		Algorithm: "HS256",
		Issuer:    "c8s-auth",
		Audience:  "c8s-api",
		Secret:    secret,
	}

	// This should work even though we called UseNoOpValidator
	err := handlerspkg.InitAuthValidator(config)
	require.NoError(t, err)
}

// TestGetUserFromContext tests user extraction from context
func TestGetUserFromContext(t *testing.T) {
	testSetup(t)
	secret := "test-secret-key-should-be-longer-32-bytes"
	config := &authpkg.Config{
		Mode:             "jwt",
		Algorithm:        "HS256",
		Issuer:           "c8s-auth",
		Audience:         "c8s-api",
		Secret:           secret,
		VerifyExpiry:     true,
		VerifySignature:  true,
		ExpiryTolerance:  0,
		DefaultNamespace: "default",
		DefaultRoles:     []string{"user"},
	}

	err := handlerspkg.InitAuthValidator(config)
	require.NoError(t, err)

	tokenString := createTestToken(secret, "c8s-auth", "c8s-api", "extract-test")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := handlerspkg.GetUserFromContext(r.Context())
		assert.True(t, ok)
		assert.NotNil(t, user)
		assert.Equal(t, "extract-test", user.ID)
		assert.Equal(t, "Test User", user.Username)
		w.WriteHeader(http.StatusOK)
	})

	handler := handlerspkg.AuthMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
