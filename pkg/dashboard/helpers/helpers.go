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

package helpers

// Dict creates a map from key-value pairs for use in templates
// Usage: {{dict "key1" value1 "key2" value2}}
func Dict(values ...interface{}) map[string]interface{} {
	if len(values)%2 != 0 {
		panic("dict: odd number of arguments")
	}
	result := make(map[string]interface{})
	for i := 0; i < len(values); i += 2 {
		key := values[i].(string)
		result[key] = values[i+1]
	}
	return result
}

// Add returns the sum of two integers
// Usage: {{add 5 3}} returns 8
func Add(a, b int) int {
	return a + b
}

// Sub returns the difference of two integers
// Usage: {{sub 10 3}} returns 7
func Sub(a, b int) int {
	return a - b
}
