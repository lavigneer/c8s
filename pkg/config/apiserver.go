package config

import (
	"flag"
	"fmt"
	"os"
)

// APIServerConfig holds all API server configuration
type APIServerConfig struct {
	// Server configuration
	Port    string
	TLSPort string
	BaseDir string

	// TLS configuration
	EnableTLS bool
	TLSCert   string
	TLSKey    string

	// Rate limiting
	APIRateLimit     float64
	APIRateLimitBurst int
	WebhookRateLimit     float64
	WebhookRateLimitBurst int

	// Request limits
	MaxRequestSize int64

	// Feature flags
	EnableMetrics bool
}

// NewAPIServerConfig creates a new API server configuration from flags and environment
func NewAPIServerConfig() *APIServerConfig {
	cfg := &APIServerConfig{
		// Defaults
		Port:              ":8080",
		TLSPort:           ":8443",
		BaseDir:           ".",
		EnableTLS:         false,
		APIRateLimit:      100.0,
		APIRateLimitBurst: 200,
		WebhookRateLimit:     10.0,
		WebhookRateLimitBurst: 20,
		MaxRequestSize:    10 * 1024 * 1024, // 10MB
		EnableMetrics:     true,
	}

	// Parse command-line flags
	flag.StringVar(&cfg.Port, "port", cfg.Port, "Port to listen on")
	flag.StringVar(&cfg.TLSPort, "tls-port", cfg.TLSPort, "TLS port to listen on")
	flag.StringVar(&cfg.BaseDir, "base-dir", cfg.BaseDir, "Base directory for templates and static files")
	flag.BoolVar(&cfg.EnableTLS, "enable-tls", cfg.EnableTLS, "Enable HTTPS/TLS")
	flag.StringVar(&cfg.TLSCert, "tls-cert", os.Getenv("TLS_CERT_PATH"), "Path to TLS certificate")
	flag.StringVar(&cfg.TLSKey, "tls-key", os.Getenv("TLS_KEY_PATH"), "Path to TLS key")
	flag.Float64Var(&cfg.APIRateLimit, "api-rate-limit", cfg.APIRateLimit, "API rate limit (requests per second)")
	flag.IntVar(&cfg.APIRateLimitBurst, "api-rate-burst", cfg.APIRateLimitBurst, "API rate limit burst")
	flag.BoolVar(&cfg.EnableMetrics, "enable-metrics", cfg.EnableMetrics, "Enable Prometheus metrics endpoint")

	return cfg
}

// Validate checks if the configuration is valid
func (c *APIServerConfig) Validate() error {
	if c.EnableTLS {
		if c.TLSCert == "" {
			return fmt.Errorf("TLS certificate path required when TLS is enabled")
		}
		if c.TLSKey == "" {
			return fmt.Errorf("TLS key path required when TLS is enabled")
		}
		if _, err := os.Stat(c.TLSCert); os.IsNotExist(err) {
			return fmt.Errorf("TLS certificate file not found: %s", c.TLSCert)
		}
		if _, err := os.Stat(c.TLSKey); os.IsNotExist(err) {
			return fmt.Errorf("TLS key file not found: %s", c.TLSKey)
		}
	}

	if c.APIRateLimit <= 0 {
		return fmt.Errorf("API rate limit must be positive")
	}

	if c.MaxRequestSize <= 0 {
		return fmt.Errorf("max request size must be positive")
	}

	return nil
}

// LogConfig logs the configuration (without sensitive values)
func (c *APIServerConfig) LogConfig() {
	fmt.Printf("API Server Configuration:\n")
	fmt.Printf("  Port: %s\n", c.Port)
	fmt.Printf("  TLS Enabled: %v\n", c.EnableTLS)
	if c.EnableTLS {
		fmt.Printf("  TLS Port: %s\n", c.TLSPort)
		fmt.Printf("  TLS Cert: %s\n", c.TLSCert)
	}
	fmt.Printf("  Base Directory: %s\n", c.BaseDir)
	fmt.Printf("  API Rate Limit: %.0f rps (burst: %d)\n", c.APIRateLimit, c.APIRateLimitBurst)
	fmt.Printf("  Max Request Size: %d MB\n", c.MaxRequestSize/(1024*1024))
	fmt.Printf("  Metrics Enabled: %v\n", c.EnableMetrics)
}
