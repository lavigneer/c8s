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

package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

var (
	logsStep      string
	logsFollow    bool
	logsTail      int
	logsAPIServer string
)

func init() {
	// These will be parsed by the logsCommand function
}

// logsCommand handles the logs subcommand
func logsCommand(args []string) error {
	fs := flag.NewFlagSet("logs", flag.ExitOnError)
	fs.StringVar(&logsStep, "step", "", "Step name to view logs from (required)")
	fs.BoolVar(&logsFollow, "follow", false, "Follow log output (stream in real-time)")
	fs.BoolVar(&logsFollow, "f", false, "Follow log output (stream in real-time) - shorthand")
	fs.IntVar(&logsTail, "tail", -1, "Number of lines to show from the end of logs")
	fs.StringVar(&logsAPIServer, "api-server", "http://localhost:8080", "API server URL")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if logsStep == "" {
		return fmt.Errorf("--step flag is required")
	}

	remainingArgs := fs.Args()
	if len(remainingArgs) != 1 {
		return fmt.Errorf("usage: c8s logs <pipelinerun-name> --step=<step-name> [--follow]")
	}

	runName := remainingArgs[0]

	if logsFollow {
		return followLogs(runName)
	}

	return fetchLogs(runName)
}

// fetchLogs retrieves logs from the API server (non-streaming)
func fetchLogs(runName string) error {
	apiURL := fmt.Sprintf("%s/api/v1/namespaces/%s/pipelineruns/%s/logs/%s",
		logsAPIServer, namespace, runName, logsStep)

	// Add tail parameter if specified
	if logsTail > 0 {
		apiURL += fmt.Sprintf("?tail=%d", logsTail)
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Stream response to stdout
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

// followLogs streams logs in real-time via WebSocket
func followLogs(runName string) error {
	// Convert HTTP URL to WebSocket URL
	wsURL, err := buildWebSocketURL(logsAPIServer, namespace, runName, logsStep)
	if err != nil {
		return fmt.Errorf("failed to build WebSocket URL: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Connecting to log stream...\n")

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer conn.Close()

	fmt.Fprintf(os.Stderr, "Connected. Streaming logs...\n\n")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	// Read messages from WebSocket and print to stdout
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				return fmt.Errorf("WebSocket error: %w", err)
			}
			// Normal close
			fmt.Fprintf(os.Stderr, "\nStream closed.\n")
			return nil
		}

		// Print log lines
		fmt.Print(string(message))
	}
}

// buildWebSocketURL converts HTTP API URL to WebSocket URL
func buildWebSocketURL(apiServer, namespace, runName, stepName string) (string, error) {
	u, err := url.Parse(apiServer)
	if err != nil {
		return "", err
	}

	// Convert scheme
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	// Build path
	u.Path = fmt.Sprintf("/api/v1/namespaces/%s/pipelineruns/%s/logs/%s",
		namespace, runName, stepName)

	// Add query parameter
	q := u.Query()
	q.Set("follow", "true")
	if logsTail > 0 {
		q.Set("tail", fmt.Sprintf("%d", logsTail))
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// tailLogs returns last N lines from a reader
func tailLogs(r io.Reader, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}

	scanner := bufio.NewScanner(r)
	lines := make([]string, 0, n)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:] // Remove oldest line
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
