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
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims for C8S authentication
type Claims struct {
	// Standard JWT claims (RFC 7519)
	Subject string   `json:"sub"`      // User ID (required)
	Name    string   `json:"name"`     // Username (required)
	Email   string   `json:"email"`    // Email address (optional)
	Audience []string `json:"aud"`     // Audience (optional but recommended)

	// C8S-specific claims
	Namespace string   `json:"namespace"` // Kubernetes namespace
	Roles     []string `json:"roles"`     // Assigned roles

	// Embed standard JWT registered claims for validation
	jwt.RegisteredClaims
}

// Valid implements jwt.Claims interface for custom validation
func (c Claims) Valid() error {
	// Let the standard claims do their validation first
	if err := c.RegisteredClaims.Valid(); err != nil {
		return err
	}

	// Validate required fields
	if c.Subject == "" {
		return jwt.NewValidationError("missing subject claim", jwt.ValidationErrorClaimsInvalid)
	}

	if c.Name == "" {
		return jwt.NewValidationError("missing name claim", jwt.ValidationErrorClaimsInvalid)
	}

	if c.Namespace == "" {
		return jwt.NewValidationError("missing namespace claim", jwt.ValidationErrorClaimsInvalid)
	}

	return nil
}

// ToUser converts JWT claims to User struct for internal use
// This allows extracting user information from the token
func (c Claims) ToUser() *User {
	roles := c.Roles
	if roles == nil {
		roles = []string{}
	}

	return &User{
		ID:        c.Subject,
		Username:  c.Name,
		Email:     c.Email,
		Namespace: c.Namespace,
		Roles:     roles,
	}
}

// User represents an authenticated user
// This is extracted from JWT claims by the validator
type User struct {
	ID        string
	Username  string
	Email     string
	Namespace string
	Roles     []string
}
