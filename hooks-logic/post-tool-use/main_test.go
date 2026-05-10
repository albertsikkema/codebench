package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleWebSearch(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer os.Unsetenv("CLAUDE_PROJECT_DIR")

	hook := HookInput{
		ToolName:     "WebSearch",
		ToolInput:    json.RawMessage(`{"query":"golang fire and forget"}`),
		ToolResponse: json.RawMessage(`{"results":[{"title":"Go concurrency","url":"https://example.com"}]}`),
		SessionID:    "test-session-1",
	}

	handleWebSearch(hook)

	// fireAndForget spawns a detached child — JSONL is written there.
	// In tests the child won't run, so we test JSONL via doProcess directly.
}

func TestHandleWebFetch(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer os.Unsetenv("CLAUDE_PROJECT_DIR")

	hook := HookInput{
		ToolName:     "WebFetch",
		ToolInput:    json.RawMessage(`{"url":"https://example.com/docs","prompt":"extract API info"}`),
		ToolResponse: json.RawMessage(`{"code":200,"result":"API documentation content"}`),
		SessionID:    "test-session-2",
	}

	handleWebFetch(hook)
}

func TestHandleWebSearchEmptyQuery(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer os.Unsetenv("CLAUDE_PROJECT_DIR")

	hook := HookInput{
		ToolName:     "WebSearch",
		ToolInput:    json.RawMessage(`{"query":""}`),
		ToolResponse: json.RawMessage(`{}`),
		SessionID:    "test-session",
	}

	handleWebSearch(hook)
	// No crash = pass (empty query is skipped)
}

func TestHandleWebFetchEmptyURL(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer os.Unsetenv("CLAUDE_PROJECT_DIR")

	hook := HookInput{
		ToolName:     "WebFetch",
		ToolInput:    json.RawMessage(`{"url":""}`),
		ToolResponse: json.RawMessage(`{}`),
		SessionID:    "test-session",
	}

	handleWebFetch(hook)
	// No crash = pass (empty URL is skipped)
}

func TestDoProcessJSONL(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "logs", "web-searches.jsonl")

	entry := `{"timestamp":"2026-03-20T14:30:00Z","session_id":"s1","query":"test"}`
	pp := ProcessPayload{
		LogFile:   logFile,
		EntryJSON: json.RawMessage(entry),
	}
	ppJSON, _ := json.Marshal(pp)

	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	w.Write(ppJSON)
	w.Close()
	os.Stdin = r

	doProcess()
	os.Stdin = oldStdin

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}

	lines := strings.TrimSpace(string(data))
	if lines != entry {
		t.Errorf("expected %q, got %q", entry, lines)
	}
}

func TestDoProcessJSONLMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "logs", "test.jsonl")

	for i := 0; i < 3; i++ {
		entry := `{"query":"test"}`
		pp := ProcessPayload{
			LogFile:   logFile,
			EntryJSON: json.RawMessage(entry),
		}
		ppJSON, _ := json.Marshal(pp)

		r, w, _ := os.Pipe()
		w.Write(ppJSON)
		w.Close()

		oldStdin := os.Stdin
		os.Stdin = r
		doProcess()
		os.Stdin = oldStdin
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 log lines, got %d", len(lines))
	}
}

func TestDoProcessAPIPost(t *testing.T) {
	var received []byte
	var receivedAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = ioutil.ReadAll(r.Body)
		receivedAPIKey = r.Header.Get("X-API-Key")
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("expected Content-Type application/json")
		}
		w.WriteHeader(201)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "logs", "test.jsonl")

	entry := json.RawMessage(`{"query":"test"}`)
	pp := ProcessPayload{
		LogFile:   logFile,
		EntryJSON: entry,
		APIURL:    server.URL,
		APIKey:    "test-key",
		APIPayload: &APIPayload{
			Type:    "web_search",
			Project: "my-project",
			Payload: entry,
		},
	}
	ppJSON, _ := json.Marshal(pp)

	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	w.Write(ppJSON)
	w.Close()
	os.Stdin = r

	doProcess()
	os.Stdin = oldStdin

	// Check JSONL was written
	if _, err := os.Stat(logFile); err != nil {
		t.Error("expected JSONL file to be created")
	}

	// Check API received the payload
	var gotPayload APIPayload
	if err := json.Unmarshal(received, &gotPayload); err != nil {
		t.Fatalf("invalid API payload: %v", err)
	}
	if gotPayload.Type != "web_search" {
		t.Errorf("expected type 'web_search', got %q", gotPayload.Type)
	}
	if gotPayload.Project != "my-project" {
		t.Errorf("expected project 'my-project', got %q", gotPayload.Project)
	}
	if receivedAPIKey != "test-key" {
		t.Errorf("expected X-API-Key 'test-key', got %q", receivedAPIKey)
	}
}

func TestDoProcessNoAPI(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "logs", "test.jsonl")

	pp := ProcessPayload{
		LogFile:   logFile,
		EntryJSON: json.RawMessage(`{"query":"test"}`),
		// No APIURL — should just write JSONL
	}
	ppJSON, _ := json.Marshal(pp)

	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	w.Write(ppJSON)
	w.Close()
	os.Stdin = r

	doProcess()
	os.Stdin = oldStdin

	if _, err := os.Stat(logFile); err != nil {
		t.Error("expected JSONL file even without API config")
	}
}

func writeEnvFile(t *testing.T, dir string, content string) {
	t.Helper()
	os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o644)
}

// clearProjectServerEnv ensures env vars don't leak between tests.
func clearProjectServerEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"PROJECT_SERVER_URL", "PROJECT_SERVER_API_KEY", "PROJECT_SERVER_PROJECT"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
}

func TestReadProjectServerConfigFromDotenv(t *testing.T) {
	clearProjectServerEnv(t)
	tmpDir := t.TempDir()
	writeEnvFile(t, tmpDir, "PROJECT_SERVER_URL=http://test-host:8080\nPROJECT_SERVER_API_KEY=my-secret-key\n")

	cfg := readProjectServerConfig(tmpDir)
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if cfg.BaseURL != "http://test-host:8080" {
		t.Errorf("expected 'http://test-host:8080', got %q", cfg.BaseURL)
	}
	if cfg.APIKey != "my-secret-key" {
		t.Errorf("expected 'my-secret-key', got %q", cfg.APIKey)
	}
	if cfg.Project != "" {
		t.Errorf("expected empty project, got %q", cfg.Project)
	}
}

func TestReadProjectServerConfigEnvOverridesDotenv(t *testing.T) {
	t.Setenv("PROJECT_SERVER_URL", "http://from-env:8080")
	t.Setenv("PROJECT_SERVER_API_KEY", "env-key")

	tmpDir := t.TempDir()
	writeEnvFile(t, tmpDir, "PROJECT_SERVER_URL=http://from-dotenv:8080\nPROJECT_SERVER_API_KEY=dotenv-key\n")

	cfg := readProjectServerConfig(tmpDir)
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if cfg.BaseURL != "http://from-env:8080" {
		t.Errorf("expected env to win, got %q", cfg.BaseURL)
	}
	if cfg.APIKey != "env-key" {
		t.Errorf("expected env key to win, got %q", cfg.APIKey)
	}
}

func TestReadProjectServerConfigTrailingSlash(t *testing.T) {
	clearProjectServerEnv(t)
	tmpDir := t.TempDir()
	writeEnvFile(t, tmpDir, "PROJECT_SERVER_URL=http://test-host:8080/\nPROJECT_SERVER_API_KEY=key\n")

	cfg := readProjectServerConfig(tmpDir)
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if cfg.BaseURL != "http://test-host:8080" {
		t.Errorf("expected trailing slash stripped, got %q", cfg.BaseURL)
	}
}

func TestReadProjectServerConfigWithProject(t *testing.T) {
	clearProjectServerEnv(t)
	tmpDir := t.TempDir()
	writeEnvFile(t, tmpDir, "PROJECT_SERVER_URL=http://test-host:8080\nPROJECT_SERVER_API_KEY=key\nPROJECT_SERVER_PROJECT=my-project\n")

	cfg := readProjectServerConfig(tmpDir)
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if cfg.Project != "my-project" {
		t.Errorf("expected 'my-project', got %q", cfg.Project)
	}
}

func TestReadProjectServerConfigQuotedValues(t *testing.T) {
	clearProjectServerEnv(t)
	tmpDir := t.TempDir()
	writeEnvFile(t, tmpDir, "PROJECT_SERVER_URL=\"http://test-host:8080\"\nPROJECT_SERVER_API_KEY='secret'\n")

	cfg := readProjectServerConfig(tmpDir)
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if cfg.BaseURL != "http://test-host:8080" {
		t.Errorf("expected quotes stripped, got %q", cfg.BaseURL)
	}
	if cfg.APIKey != "secret" {
		t.Errorf("expected quotes stripped, got %q", cfg.APIKey)
	}
}

func TestReadProjectServerConfigMissing(t *testing.T) {
	clearProjectServerEnv(t)
	cfg := readProjectServerConfig(t.TempDir())
	if cfg != nil {
		t.Error("expected nil config when no .env and no env vars")
	}
}

func TestFireAndForgetNoProjectDir(t *testing.T) {
	os.Unsetenv("CLAUDE_PROJECT_DIR")
	fireAndForget("web_search", "test.jsonl", []byte(`{"query":"test"}`))
	// No crash = pass
}

func TestFireAndForgetNoMcpJson(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer os.Unsetenv("CLAUDE_PROJECT_DIR")

	fireAndForget("web_search", "test.jsonl", []byte(`{"query":"test"}`))
	// No crash = pass (still spawns child for JSONL, just no API)
}
