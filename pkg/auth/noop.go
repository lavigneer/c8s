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

import "fmt"

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
			Mode:             ModeNone,
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
