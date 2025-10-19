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

package handlers

import (
	"net/http"
	"strings"
)

// extractNamespace extracts the namespace from the request URL path
// Expected pattern: /api/v1/namespaces/{namespace}/...
func extractNamespace(r *http.Request) string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for i, part := range parts {
		if part == "namespaces" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// extractResourceName extracts the resource name from the request URL path
// Expected pattern: .../pipelineconfigs/{name} or .../pipelineruns/{name}
func extractResourceName(r *http.Request) string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) > 0 {
		// The last part is the resource name
		lastPart := parts[len(parts)-1]
		// Check if this is a collection endpoint (ends with resource type)
		if lastPart == "pipelineconfigs" || lastPart == "pipelineruns" ||
			lastPart == "repositoryconnections" || lastPart == "logs" {
			return ""
		}
		return lastPart
	}
	return ""
}

// extractStepName extracts the step name from logs endpoint
// Expected pattern: .../pipelineruns/{name}/logs/{step}
func extractStepName(r *http.Request) string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for i, part := range parts {
		if part == "logs" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
