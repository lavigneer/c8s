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

package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authpkg "github.com/org/c8s/pkg/auth"
)

// TestHS256TokenValidation tests basic HS256 token validation
func TestHS256TokenValidation(t *testing.T) {
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

	validator, err := authpkg.NewValidator(config)
	require.NoError(t, err)

	// Create a valid token
	now := time.Now()
	claims := authpkg.Claims{
		Subject:   "user-123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Namespace: "default",
		Roles:     []string{"admin", "viewer"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "c8s-auth",
			Audience:  jwt.ClaimStrings{"c8s-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	// Test valid token
	validClaims, err := validator.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", validClaims.Subject)
	assert.Equal(t, "John Doe", validClaims.Name)
	assert.Equal(t, "john@example.com", validClaims.Email)
	assert.Equal(t, "default", validClaims.Namespace)
	assert.Equal(t, []string{"admin", "viewer"}, validClaims.Roles)
}

// TestExpiredToken tests that expired tokens are rejected
func TestExpiredToken(t *testing.T) {
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

	validator, err := authpkg.NewValidator(config)
	require.NoError(t, err)

	// Create an expired token
	now := time.Now()
	claims := authpkg.Claims{
		Subject:   "user-123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Namespace: "default",
		Roles:     []string{"user"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "c8s-auth",
			Audience:  jwt.ClaimStrings{"c8s-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	// Test expired token
	_, err = validator.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// TestInvalidSignature tests that tokens with invalid signatures are rejected
func TestInvalidSignature(t *testing.T) {
	secret := "test-secret-key-should-be-longer-32-bytes"
	wrongSecret := "wrong-secret-key-should-be-longer-32"

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

	validator, err := authpkg.NewValidator(config)
	require.NoError(t, err)

	// Create a token with a different secret
	now := time.Now()
	claims := authpkg.Claims{
		Subject:   "user-123",
		Name:      "John Doe",
		Namespace: "default",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "c8s-auth",
			Audience:  jwt.ClaimStrings{"c8s-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(wrongSecret))
	require.NoError(t, err)

	// Test invalid signature
	_, err = validator.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature")
}

// TestMissingRequiredClaims tests that tokens missing required claims are rejected
func TestMissingRequiredClaims(t *testing.T) {
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

	validator, err := authpkg.NewValidator(config)
	require.NoError(t, err)

	now := time.Now()

	testCases := []struct {
		name   string
		claims authpkg.Claims
		errMsg string
	}{
		{
			name: "missing subject",
			claims: authpkg.Claims{
				Subject:   "", // Missing
				Name:      "John Doe",
				Namespace: "default",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "c8s-auth",
					ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
				},
			},
			errMsg: "subject",
		},
		{
			name: "missing name",
			claims: authpkg.Claims{
				Subject:   "user-123",
				Name:      "", // Missing
				Namespace: "default",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "c8s-auth",
					ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
				},
			},
			errMsg: "name",
		},
		{
			name: "missing namespace",
			claims: authpkg.Claims{
				Subject:   "user-123",
				Name:      "John Doe",
				Namespace: "", // Missing
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "c8s-auth",
					ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
				},
			},
			errMsg: "namespace",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, tc.claims)
			tokenString, err := token.SignedString([]byte(secret))
			require.NoError(t, err)

			_, err = validator.ValidateToken(tokenString)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

// TestClaimsToUser tests conversion of claims to user struct
func TestClaimsToUser(t *testing.T) {
	claims := authpkg.Claims{
		Subject:   "user-123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Namespace: "production",
		Roles:     []string{"admin", "viewer"},
	}

	user := claims.ToUser()

	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "John Doe", user.Username)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, "production", user.Namespace)
	assert.Equal(t, []string{"admin", "viewer"}, user.Roles)
}

// TestValidateTokenAndGetUser tests full validation flow
func TestValidateTokenAndGetUser(t *testing.T) {
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
		DefaultRoles:     []string{"viewer"},
	}

	validator, err := authpkg.NewValidator(config)
	require.NoError(t, err)

	now := time.Now()
	claims := authpkg.Claims{
		Subject:   "user-456",
		Name:      "Jane Smith",
		Email:     "jane@example.com",
		Namespace: "staging",
		Roles:     []string{"developer"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "c8s-auth",
			Audience:  jwt.ClaimStrings{"c8s-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	user, err := validator.ValidateTokenAndGetUser(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, "user-456", user.ID)
	assert.Equal(t, "Jane Smith", user.Username)
	assert.Equal(t, "staging", user.Namespace)
	assert.Equal(t, []string{"developer"}, user.Roles)
}

// TestEmptyToken tests that empty tokens are rejected
func TestEmptyToken(t *testing.T) {
	config := &authpkg.Config{
		Mode:             "jwt",
		Algorithm:        "HS256",
		Issuer:           "c8s-auth",
		Audience:         "c8s-api",
		Secret:           "test-secret-key-should-be-longer-32-bytes",
		VerifyExpiry:     true,
		VerifySignature:  true,
		DefaultNamespace: "default",
		DefaultRoles:     []string{"user"},
	}

	validator, err := authpkg.NewValidator(config)
	require.NoError(t, err)

	_, err = validator.ValidateToken("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestNoOpValidator tests the development-only validator
func TestNoOpValidator(t *testing.T) {
	validator := authpkg.NewNoOpValidator()

	// Any non-empty token should be accepted
	claims, err := validator.ValidateToken("any-token-string")
	assert.NoError(t, err)
	assert.Equal(t, "dev-user", claims.Subject)

	// Empty token should fail
	_, err = validator.ValidateToken("")
	assert.Error(t, err)

	// Test getting user
	user, err := validator.ValidateTokenAndGetUser("any-token")
	assert.NoError(t, err)
	assert.Equal(t, "dev-user", user.ID)
	assert.Equal(t, "Developer", user.Username)
}

// TestInvalidIssuer tests that tokens from wrong issuer are rejected
func TestInvalidIssuer(t *testing.T) {
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

	validator, err := authpkg.NewValidator(config)
	require.NoError(t, err)

	now := time.Now()
	claims := authpkg.Claims{
		Subject:   "user-123",
		Name:      "John Doe",
		Namespace: "default",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wrong-issuer", // Wrong issuer
			Audience:  jwt.ClaimStrings{"c8s-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	_, err = validator.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "issuer")
}

// TestInvalidAudience tests that tokens with wrong audience are rejected
func TestInvalidAudience(t *testing.T) {
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

	validator, err := authpkg.NewValidator(config)
	require.NoError(t, err)

	now := time.Now()
	claims := authpkg.Claims{
		Subject:   "user-123",
		Name:      "John Doe",
		Namespace: "default",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "c8s-auth",
			Audience:  jwt.ClaimStrings{"wrong-api"}, // Wrong audience
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	_, err = validator.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audience")
}
