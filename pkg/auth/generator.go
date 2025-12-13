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
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenGenerator handles JWT token generation
type TokenGenerator struct {
	config *Config
}

// NewTokenGenerator creates a new JWT token generator
func NewTokenGenerator(config *Config) (*TokenGenerator, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid auth config: %w", err)
	}

	return &TokenGenerator{config: config}, nil
}

// GenerateToken creates a new JWT token for the given user
func (g *TokenGenerator) GenerateToken(userID, username, email, namespace string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	claims := &Claims{
		Subject:   userID,
		Name:      username,
		Email:     email,
		Namespace: namespace,
		Roles:     g.config.DefaultRoles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    g.config.Issuer,
			Audience:  jwt.ClaimStrings{g.config.Audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(g.config.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
