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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	// AlgorithmHS256 represents the HMAC-SHA256 signing algorithm
	AlgorithmHS256 = "HS256"
	// AlgorithmRS256 represents the RSA-SHA256 signing algorithm
	AlgorithmRS256 = "RS256"
)

// Config holds authentication configuration
type Config struct {
	// Authentication mode: "jwt", "none" (dev only)
	Mode string

	// JWT Configuration
	Algorithm string // "HS256" or "RS256"
	Issuer    string // Expected issuer
	Audience  string // Expected audience

	// For HS256 (symmetric key)
	Secret string // Base64-encoded secret

	// For RS256 (asymmetric key)
	PublicKeyPath string // Path to PEM public key file
	JWKSUrl       string // JWKS endpoint URL (alternative to file)

	// Token validation
	VerifyExpiry    bool
	VerifySignature bool
	ExpiryTolerance time.Duration

	// Default values
	DefaultNamespace string
	DefaultRoles     []string
}

// NewConfigFromEnv loads authentication configuration from environment variables
func NewConfigFromEnv() *Config {
	roles := strings.Split(os.Getenv("JWT_DEFAULT_ROLES"), ",")
	if len(roles) == 1 && roles[0] == "" {
		roles = []string{"user"}
	}

	return &Config{
		Mode:             getEnv("AUTH_MODE", "jwt"),
		Algorithm:        getEnv("JWT_ALGORITHM", AlgorithmHS256),
		Issuer:           getEnv("JWT_ISSUER", "c8s-auth"),
		Audience:         getEnv("JWT_AUDIENCE", "c8s-api"),
		Secret:           os.Getenv("JWT_SECRET"),
		PublicKeyPath:    getEnv("JWT_PUBLIC_KEY_PATH", ""),
		JWKSUrl:          getEnv("JWT_JWKS_URL", ""),
		VerifyExpiry:     getEnvBool("JWT_VERIFY_EXPIRY", true),
		VerifySignature:  getEnvBool("JWT_VERIFY_SIGNATURE", true),
		ExpiryTolerance:  getEnvDuration("JWT_EXPIRY_TOLERANCE", 60*time.Second),
		DefaultNamespace: getEnv("JWT_DEFAULT_NAMESPACE", "default"),
		DefaultRoles:     roles,
	}
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	if c.Mode != "jwt" && c.Mode != "none" {
		return fmt.Errorf("invalid AUTH_MODE: %s (must be 'jwt' or 'none')", c.Mode)
	}

	if c.Mode == "none" {
		return nil // Skip validation for dev-only mode
	}

	if c.Algorithm != AlgorithmHS256 && c.Algorithm != AlgorithmRS256 {
		return fmt.Errorf("invalid JWT_ALGORITHM: %s (must be 'HS256' or 'RS256')", c.Algorithm)
	}

	if c.Algorithm == AlgorithmHS256 && c.Secret == "" {
		return fmt.Errorf("JWT_SECRET required for HS256 algorithm")
	}

	if c.Algorithm == AlgorithmRS256 && c.PublicKeyPath == "" && c.JWKSUrl == "" {
		return fmt.Errorf("either JWT_PUBLIC_KEY_PATH or JWT_JWKS_URL required for RS256 algorithm")
	}

	if c.Issuer == "" {
		return fmt.Errorf("JWT_ISSUER required")
	}

	if c.Audience == "" {
		return fmt.Errorf("JWT_AUDIENCE required")
	}

	return nil
}

// LoadPublicKey loads and parses RSA public key from PEM file
func (c *Config) LoadPublicKey() (*rsa.PublicKey, error) {
	if c.Algorithm != AlgorithmRS256 {
		return nil, fmt.Errorf("LoadPublicKey called for non-RS256 algorithm")
	}

	if c.PublicKeyPath == "" {
		return nil, fmt.Errorf("PublicKeyPath not set")
	}

	pemData, err := os.ReadFile(c.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}

	return rsaKey, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val == "true" || val == "1" || val == "yes"
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return defaultValue
	}
	return d
}
