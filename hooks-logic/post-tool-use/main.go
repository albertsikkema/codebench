package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Input from Claude Code hook
type HookInput struct {
	ToolName     string          `json:"tool_name"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
	SessionID    string          `json:"session_id"`
}

// Parsed tool inputs
type WebSearchInput struct {
	Query string `json:"query"`
}

type WebFetchInput struct {
	URL    string `json:"url"`
	Prompt string `json:"prompt"`
}

// JSONL log entries
type SearchLogEntry struct {
	Timestamp string          `json:"timestamp"`
	SessionID string          `json:"session_id"`
	Query     string          `json:"query"`
	Results   json.RawMessage `json:"results"`
}

type FetchLogEntry struct {
	Timestamp string          `json:"timestamp"`
	SessionID string          `json:"session_id"`
	URL       string          `json:"url"`
	Prompt    string          `json:"prompt"`
	Status    json.RawMessage `json:"status"`
	Result    json.RawMessage `json:"result"`
}

// API payload wraps the log entry with type and project info
type APIPayload struct {
	Type    string          `json:"type"`
	Project string          `json:"project"`
	Payload json.RawMessage `json:"payload"`
}

// projectServerConfig holds the project-server URL, API key, and project name
type projectServerConfig struct {
	BaseURL string
	APIKey  string
	Project string
}

// debug is enabled when CLAUDE_HOOKS_DEBUG is set to a truthy value.
var debug = func() bool {
	v := strings.ToLower(os.Getenv("CLAUDE_HOOKS_DEBUG"))
	return v == "1" || v == "true" || v == "yes"
}()

// debugLog writes to stderr when CLAUDE_HOOKS_DEBUG is set.
func debugLog(msg string) {
	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG post-tool-use] %s\n", msg)
	}
}

// parseDotenv reads a .env file and returns a map of key-value pairs.
func parseDotenv(projectDir string) map[string]string {
	vars := make(map[string]string)
	data, err := os.ReadFile(filepath.Join(projectDir, ".env"))
	if err != nil {
		return vars
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		// Strip matching outer quotes
		if len(v) >= 2 && v[0] == v[len(v)-1] && (v[0] == '"' || v[0] == '\'') {
			v = v[1 : len(v)-1]
		}
		vars[k] = v
	}
	return vars
}

// envOrDotenv returns the env var if set, otherwise the .env value.
func envOrDotenv(key string, dotenv map[string]string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return dotenv[key]
}

// readProjectServerConfig reads project-server configuration from .env
// (with env var overrides): PROJECT_SERVER_URL, PROJECT_SERVER_API_KEY,
// and PROJECT_SERVER_PROJECT.
func readProjectServerConfig(projectDir string) *projectServerConfig {
	dotenv := parseDotenv(projectDir)

	baseURL := envOrDotenv("PROJECT_SERVER_URL", dotenv)
	if baseURL == "" {
		debugLog("PROJECT_SERVER_URL not set; API logging disabled (local JSONL still written)")
		return nil
	}

	return &projectServerConfig{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  envOrDotenv("PROJECT_SERVER_API_KEY", dotenv),
		Project: envOrDotenv("PROJECT_SERVER_PROJECT", dotenv),
	}
}

func main() {
	// Self-invocation mode: --process reads a ProcessPayload from stdin,
	// writes JSONL, and optionally POSTs to the API
	if len(os.Args) >= 2 && os.Args[1] == "--process" {
		doProcess()
		return
	}

	input, err := io.ReadAll(os.Stdin)
	if err != nil || len(input) == 0 {
		os.Exit(0)
	}

	var hook HookInput
	if err := json.Unmarshal(input, &hook); err != nil {
		os.Exit(0)
	}

	switch hook.ToolName {
	case "WebSearch":
		handleWebSearch(hook)
	case "WebFetch":
		handleWebFetch(hook)
	}
}

func handleWebSearch(hook HookInput) {
	var si WebSearchInput
	if err := json.Unmarshal(hook.ToolInput, &si); err != nil || si.Query == "" {
		return
	}

	// Extract results from tool_response
	var resp map[string]json.RawMessage
	if err := json.Unmarshal(hook.ToolResponse, &resp); err != nil {
		return
	}

	entry := SearchLogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SessionID: hook.SessionID,
		Query:     si.Query,
		Results:   resp["results"],
	}

	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return
	}

	fireAndForget("web_search", "web-searches.jsonl", entryJSON)
}

func handleWebFetch(hook HookInput) {
	var fi WebFetchInput
	if err := json.Unmarshal(hook.ToolInput, &fi); err != nil || fi.URL == "" {
		return
	}

	// Extract status and result from tool_response
	var resp map[string]json.RawMessage
	if err := json.Unmarshal(hook.ToolResponse, &resp); err != nil {
		return
	}

	entry := FetchLogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SessionID: hook.SessionID,
		URL:       fi.URL,
		Prompt:    fi.Prompt,
		Status:    resp["code"],
		Result:    resp["result"],
	}

	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return
	}

	fireAndForget("web_fetch", "web-fetches.jsonl", entryJSON)
}


// ProcessPayload is what the parent sends to the detached child process
type ProcessPayload struct {
	LogFile    string          `json:"log_file"`
	EntryJSON  json.RawMessage `json:"entry_json"`
	APIPayload *APIPayload     `json:"api_payload,omitempty"`
	APIURL     string          `json:"api_url,omitempty"`
	APIKey     string          `json:"api_key,omitempty"`
}

func fireAndForget(entryType string, logFile string, entryJSON []byte) {
	projectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	if projectDir == "" {
		debugLog(fmt.Sprintf("%s skipped: CLAUDE_PROJECT_DIR not set", entryType))
		return
	}

	pp := ProcessPayload{
		LogFile:   filepath.Join(projectDir, ".claude", "workspace", "logs", logFile),
		EntryJSON: entryJSON,
	}

	// Add API config if available
	cfg := readProjectServerConfig(projectDir)
	if cfg != nil {
		pp.APIURL = cfg.BaseURL + "/api/web-logs"
		pp.APIKey = cfg.APIKey
		pp.APIPayload = &APIPayload{
			Type:    entryType,
			Project: cfg.Project,
			Payload: entryJSON,
		}
		debugLog(fmt.Sprintf("%s: logging to %s and POST %s", entryType, pp.LogFile, pp.APIURL))
	} else {
		debugLog(fmt.Sprintf("%s: logging to %s (API disabled)", entryType, pp.LogFile))
	}

	ppJSON, err := json.Marshal(pp)
	if err != nil {
		debugLog(fmt.Sprintf("%s skipped: marshal payload: %v", entryType, err))
		return
	}

	// Re-invoke self with --process flag, detached from parent process
	self, err := os.Executable()
	if err != nil {
		debugLog(fmt.Sprintf("%s skipped: os.Executable: %v", entryType, err))
		return
	}

	cmd := exec.Command(self, "--process")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		debugLog(fmt.Sprintf("%s skipped: StdinPipe: %v", entryType, err))
		return
	}

	if err := cmd.Start(); err != nil {
		debugLog(fmt.Sprintf("%s skipped: cmd.Start: %v", entryType, err))
		return
	}

	// Write payload and close stdin so the child can read it fully
	stdin.Write(ppJSON)
	stdin.Close()
	// Don't call cmd.Wait() — fire and forget
}

// doProcess is the detached child: writes JSONL and optionally POSTs to the API.
func doProcess() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil || len(data) == 0 {
		os.Exit(1)
	}

	var pp ProcessPayload
	if err := json.Unmarshal(data, &pp); err != nil {
		fmt.Fprintf(os.Stderr, "web-log process: invalid payload: %v\n", err)
		os.Exit(1)
	}

	// Write JSONL
	if pp.LogFile != "" {
		os.MkdirAll(filepath.Dir(pp.LogFile), 0o755)
		f, err := os.OpenFile(pp.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			f.Write(pp.EntryJSON)
			f.Write([]byte("\n"))
			f.Close()
		}
	}

	// POST to API
	if pp.APIURL != "" && pp.APIPayload != nil {
		body, err := json.Marshal(pp.APIPayload)
		if err != nil {
			return
		}

		req, err := http.NewRequest("POST", pp.APIURL, bytes.NewReader(body))
		if err != nil {
			fmt.Fprintf(os.Stderr, "web-log post failed: %v\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if pp.APIKey != "" {
			req.Header.Set("X-API-Key", pp.APIKey)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "web-log post failed: %v\n", err)
			return
		}
		resp.Body.Close()
	}
}

