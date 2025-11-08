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
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Validator handles JWT token validation
type Validator struct {
	config      *Config
	publicKey   *rsa.PublicKey // For RS256
	lastKeyLoad time.Time
}

// NewValidator creates a new JWT validator with the given configuration
func NewValidator(config *Config) (*Validator, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid auth config: %w", err)
	}

	v := &Validator{
		config: config,
	}

	// For RS256, load the public key upfront
	if config.Algorithm == "RS256" && config.PublicKeyPath != "" {
		key, err := config.LoadPublicKey()
		if err != nil {
			return nil, fmt.Errorf("failed to load public key: %w", err)
		}
		v.publicKey = key
		v.lastKeyLoad = time.Now()
	}

	return v, nil
}

// ValidateToken validates a JWT token string and returns claims if valid
func (v *Validator) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("empty token")
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, v.keyFunc)
	if err != nil {
		return nil, v.formatError(err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	// Validate required claims
	if err := claims.Valid(); err != nil {
		return nil, err
	}

	// Perform additional validation
	if err := v.validateClaims(claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// ValidateTokenAndGetUser validates a token and converts claims to User
func (v *Validator) ValidateTokenAndGetUser(tokenString string) (*User, error) {
	claims, err := v.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Apply defaults if necessary
	user := claims.ToUser()
	if user.Namespace == "" {
		user.Namespace = v.config.DefaultNamespace
	}
	if user.Roles == nil || len(user.Roles) == 0 {
		user.Roles = v.config.DefaultRoles
	}

	return user, nil
}

// keyFunc is the callback function for jwt.ParseWithClaims to get the signing key
func (v *Validator) keyFunc(token *jwt.Token) (interface{}, error) {
	// Verify the algorithm
	if token.Method.Alg() != v.config.Algorithm {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	switch v.config.Algorithm {
	case "HS256":
		return []byte(v.config.Secret), nil
	case "RS256":
		if v.publicKey == nil {
			return nil, fmt.Errorf("public key not loaded")
		}
		return v.publicKey, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", v.config.Algorithm)
	}
}

// validateClaims performs additional validation on the claims
func (v *Validator) validateClaims(claims *Claims) error {
	if !v.config.VerifySignature {
		// Note: Signature was already verified by jwt.ParseWithClaims
		// This flag is more for documentation/config purposes
	}

	// Check expiration if enabled
	if v.config.VerifyExpiry {
		now := time.Now().Unix()
		if claims.ExpiresAt != nil && now > claims.ExpiresAt.Unix()+int64(v.config.ExpiryTolerance.Seconds()) {
			return fmt.Errorf("token has expired")
		}
	}

	// Validate issuer
	if claims.Issuer != "" && claims.Issuer != v.config.Issuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", v.config.Issuer, claims.Issuer)
	}

	// Validate audience - check if the expected audience is in the token's audience list
	// Note: JWT library stores audience in RegisteredClaims.Audience, not our custom Audience field
	if len(claims.RegisteredClaims.Audience) > 0 {
		found := false
		for _, aud := range claims.RegisteredClaims.Audience {
			if aud == v.config.Audience {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid audience: expected %s, got %v", v.config.Audience, claims.RegisteredClaims.Audience)
		}
	}

	return nil
}

// formatError formats JWT parsing errors into user-friendly messages
func (v *Validator) formatError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Handle common JWT errors
	if strings.Contains(errStr, "signature is invalid") {
		return fmt.Errorf("invalid token signature")
	}
	if strings.Contains(errStr, "token is expired") {
		return fmt.Errorf("token has expired")
	}
	if strings.Contains(errStr, "claims is invalid") {
		return fmt.Errorf("invalid token claims")
	}
	if strings.Contains(errStr, "malformed") {
		return fmt.Errorf("malformed token")
	}

	// For any other error, return generic message to avoid leaking details
	return fmt.Errorf("invalid token")
}

// NoOpValidator is a validator that accepts any token (dev-only)
// Should never be used in production
type NoOpValidator struct {
	config *Config
}

// NewNoOpValidator creates a validator that accepts all tokens
// WARNING: This is for development only!
func NewNoOpValidator() *NoOpValidator {
	return &NoOpValidator{
		config: &Config{
			Mode:             "none",
			DefaultNamespace: "default",
			DefaultRoles:     []string{"user"},
		},
	}
}

// ValidateToken accepts any non-empty token
func (n *NoOpValidator) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("empty token")
	}

	// Return a default claims object - DO NOT USE IN PRODUCTION
	return &Claims{
		Subject:   "dev-user",
		Name:      "Developer",
		Email:     "dev@localhost",
		Namespace: "default",
		Roles:     []string{"admin"},
	}, nil
}

// ValidateTokenAndGetUser returns a development user for any token
func (n *NoOpValidator) ValidateTokenAndGetUser(tokenString string) (*User, error) {
	claims, err := n.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	return claims.ToUser(), nil
}
