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

package storage

import "os"

// NewConfigFromEnv creates a storage Config from environment variables
// Expected environment variables:
// - S3_BUCKET: S3 bucket name (required for S3 mode)
// - S3_ENDPOINT: S3 endpoint URL (e.g., http://minio.c8s-system.svc.cluster.local:9000)
// - S3_REGION: AWS region (defaults to us-east-1)
// - AWS_ACCESS_KEY_ID: S3 access key
// - AWS_SECRET_ACCESS_KEY: S3 secret key
// - S3_USE_PATH_STYLE: Use path-style URLs for S3-compatible services (defaults to true)
func NewConfigFromEnv() *Config {
	return &Config{
		Bucket:          os.Getenv("S3_BUCKET"),
		Region:          getEnvOrDefault("S3_REGION", "us-east-1"),
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		UsePathStyle:    getEnvOrDefault("S3_USE_PATH_STYLE", "true") == "true",
	}
}

// getEnvOrDefault returns the environment variable value or a default if not set
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
