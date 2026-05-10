package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// --- Input/Output types ---

type hookInput struct {
	ToolName  string         `json:"tool_name"`
	ToolInput map[string]any `json:"tool_input"`
}

type hookOutput struct {
	HookSpecificOutput hookSpecific `json:"hookSpecificOutput"`
}

type hookSpecific struct {
	HookEventName            string         `json:"hookEventName"`
	PermissionDecision       string         `json:"permissionDecision"`
	PermissionDecisionReason string         `json:"permissionDecisionReason,omitempty"`
	UpdatedInput             map[string]any `json:"updatedInput,omitempty"`
}

// --- Environment ---

var (
	debug         = envBool("CLAUDE_HOOKS_DEBUG")
	containerMode = envBool("CLAUDE_CONTAINER_MODE")
	projectDir    = envOr("CLAUDE_PROJECT_DIR", ".")
	sessionID     = os.Getenv("CLAUDE_SESSION_ID")
)

func envBool(key string) bool {
	v := strings.ToLower(os.Getenv(key))
	return v == "1" || v == "true"
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// --- Compiled regex patterns ---

// Each category is a slice compiled at init time.

var dangerousRmPatterns = compileAll(
	`\brm\s+.*-[a-z]*r[a-z]*f`,
	`\brm\s+.*-[a-z]*f[a-z]*r`,
	`\brm\s+--recursive\s+--force`,
	`\brm\s+--force\s+--recursive`,
	`\brm\s+-r\s+.*-f`,
	`\brm\s+-f\s+.*-r`,
)

var dangerousRmPathPatterns = compileAll(
	`\s/$`,
	`\s/\*`,
	`\s~/?`,
	`(?i)\s\$HOME`,
	`\s\.\./?`,
	`\s\.$`,
)

var rmRecursivePattern = regexp.MustCompile(`\brm\s+.*-[a-z]*r`)

var forkBombPatterns = compileAll(
	`:\(\)\s*\{\s*:\|:&\s*\}\s*;:`,
	`\./\w+\s*&\s*\./\w+`,
	`(?i)while\s+true.*fork`,
	`(?i)fork\s*\(\s*\)\s*while`,
)

var gitPushMainPattern = regexp.MustCompile(`git\s+push\b.*\s(main|master)\b`)
var gitPushForcePattern = regexp.MustCompile(`git\s+push\b.*--force`)
var gitPushForceLease = regexp.MustCompile(`--force-with-lease`)
var gitPushDashF = regexp.MustCompile(`git\s+push\b.*\s-f\b`)
var gitResetHard = regexp.MustCompile(`git\s+reset\s+--hard`)
var gitCleanFd = regexp.MustCompile(`git\s+clean\s+-f`)

var dangerousDiskPatterns = compileAll(
	`\bdd\s+.*of=/dev/`,
	`\bmkfs`,
	`>\s*/dev/sd`,
)

var dangerousDeletePatterns = compileAll(
	`\bfind\s+.*-delete\b`,
)

var cloudMetadataPatterns = compileAll(
	`\bcurl\s+.*http://169\.254\.169\.254`,
	`\bwget\s+.*http://169\.254\.169\.254`,
	`\bcurl\s+.*http://metadata\.google\.internal`,
	`\bwget\s+.*http://metadata\.google\.internal`,
)

var pipeToShellPatterns = compileAll(
	`\bcurl\s+.*\|\s*(sh|bash)\b`,
	`\bwget\s+.*\|\s*(sh|bash)\b`,
	`\|\s*base64\s+-d\s*\|\s*(sh|bash)\b`,
)

var reverseShellPatterns = compileAll(
	`bash\s+-i\s+>&\s*/dev/tcp/`,
	`\|\s*nc\s+-l`,
	`\bnc\s+.*-e\s+/bin/(ba)?sh`,
	`(?i)\bpython[23]?\s+-c\s+.*socket.*connect`,
)

var containerEscapePatterns = compileAll(
	`\bnsenter\b`,
	`\bdocker\s+(container\s+)?run\s+.*--privileged`,
	`\bdocker\s+(container\s+)?run\s+.*-v\s+/:/`,
	`\bdocker\s+(container\s+)?run\s+.*--pid=host`,
)

var credentialExfilPatterns = compileAll(
	`\benv\s*\|\s*base64\b`,
	`\bprintenv\s*\|\s*base64\b`,
	`\bcat\s+/etc/shadow\b`,
)

// bashSensitivePathPatterns catches cat/less/head/tail of sensitive paths
// that would otherwise only be checked for file tool operations.
var bashSensitivePathPatterns = compileAll(
	`\b(cat|less|head|tail|more)\s+.*\.(pem|key|p12|pfx)\b`,
	`\b(cat|less|head|tail|more)\s+.*\.aws/credentials`,
	`\b(cat|less|head|tail|more)\s+.*\.ssh/`,
	`\b(cat|less|head|tail|more)\s+.*\.kube/config`,
	`\b(cat|less|head|tail|more)\s+.*\.gnupg/`,
	`\b(cat|less|head|tail|more)\s+.*\.docker/config\.json`,
	`\b(cat|less|head|tail|more)\s+.*\.netrc`,
	`\b(cat|less|head|tail|more)\s+.*\.npmrc`,
	`\b(cat|less|head|tail|more)\s+.*\.pypirc`,
)

var cryptoMiningPatterns = compileAll(
	`(?i)xmrig`,
)

var envAccessPatterns = compileAll(
	`\bcat\s+[^|]*\.env\b`,
	`\bless\s+[^|]*\.env\b`,
	`\bhead\s+[^|]*\.env\b`,
	`\btail\s+[^|]*\.env\b`,
	`>\s*[^\s]*\.env\b`,
	`\bcp\s+[^|]*\.env\b`,
	`\bmv\s+[^|]*\.env\b`,
	`\bsource\s+[^|]*\.env\b`,
	`\.\s+[^|]*\.env\b`,
)

var envSafePattern = regexp.MustCompile(`\.env\.(sample|example|template)\b`)

// envRefPattern finds each .env reference and captures what follows it
var envRefPattern = regexp.MustCompile(`\.env\b(\.\w+)?`)

// hasUnsafeEnvRef checks if any .env reference in the command is NOT a safe variant.
// "cat .env" → true (unsafe), "cat .env.example" → false (safe),
// "cat .env.example .env" → true (second ref is unsafe).
func hasUnsafeEnvRef(cmd string) bool {
	for _, match := range envRefPattern.FindAllString(cmd, -1) {
		if !envSafePattern.MatchString(match) {
			return true
		}
	}
	return false
}

var sensitiveFilePatterns = []struct {
	re   *regexp.Regexp
	desc string
}{
	{regexp.MustCompile(`\.pem$`), "PEM certificate/key file"},
	{regexp.MustCompile(`\.key$`), "Key file"},
	{regexp.MustCompile(`\.p12$`), "PKCS12 certificate"},
	{regexp.MustCompile(`\.pfx$`), "PFX certificate"},
	{regexp.MustCompile(`credentials\.(json|yaml|yml|xml|ini|conf)$`), "Credentials file"},
	{regexp.MustCompile(`secrets?\.(json|yaml|yml|xml|ini|conf)$`), "Secrets file"},
	{regexp.MustCompile(`\.kube/config`), "Kubernetes config"},
	{regexp.MustCompile(`\.aws/credentials`), "AWS credentials"},
	{regexp.MustCompile(`\.ssh/`), "SSH directory"},
	{regexp.MustCompile(`\.gnupg/`), "GPG directory"},
	{regexp.MustCompile(`\.netrc`), "Netrc file"},
	{regexp.MustCompile(`\.npmrc`), "NPM config with tokens"},
	{regexp.MustCompile(`\.pypirc`), "PyPI config with tokens"},
	{regexp.MustCompile(`\.vault-token$`), "Vault token"},
	{regexp.MustCompile(`\.htpasswd$`), "Apache password file"},
	{regexp.MustCompile(`\.pgpass$`), "PostgreSQL password file"},
	{regexp.MustCompile(`\.my\.cnf$`), "MySQL config file"},
	{regexp.MustCompile(`\.docker/config\.json$`), "Docker registry auth"},
	{regexp.MustCompile(`service[_-]account.*\.json$`), "GCP service account"},
	{regexp.MustCompile(`\.tfstate$`), "Terraform state file"},
}

var sensitiveFiles = map[string]bool{
	".env":             true,
	".env.local":       true,
	".env.production":  true,
	".env.development": true,
	".env.staging":     true,
	".envrc":           true,
}

var allowedEnvFiles = map[string]bool{
	".env.sample":   true,
	".env.example":  true,
	".env.template": true,
}

// --- GTK rewrite ---

var gtkPrefixes = []string{
	"git ", "cargo ", "go test", "go build", "go vet", "golangci-lint ",
	"gh ", "docker ", "kubectl ", "pnpm ", "npm ", "npx ",
	"vitest ", "playwright ", "pytest ", "ruff ", "mypy ", "pip ",
	"dotnet ", "eslint ", "tsc ", "prettier ", "next ", "prisma ",
	"curl ", "ls ", "find ", "grep ",
}

// --- Helpers ---

func compileAll(patterns ...string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		out[i] = regexp.MustCompile(p)
	}
	return out
}

func debugLog(msg string) {
	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s\n", msg)
	}
}

func normalize(cmd string) string {
	return strings.Join(strings.Fields(strings.ToLower(cmd)), " ")
}

// splitSegments splits a command on shell separators (&&, ;, ||)
// so each segment can be checked independently. This prevents false positives
// from unrelated commands in a chain. For example:
//
//	"git push origin && git checkout main" → allowed
//	  (push has no main target, checkout is a separate segment)
//
//	"git push origin main && echo done" → blocked
//	  (the push segment targets main)
//
// Does not handle quoted strings — "git commit -m 'fix; cleanup'" would
// split on the semicolon inside quotes. This errs on the side of blocking.
func splitSegments(cmd string) []string {
	var segments []string
	var buf strings.Builder
	var quote rune // 0 = not in quotes, '\'' or '"' = in that quote type
	runes := []rune(cmd)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		// Track quote state — skip separators inside quotes
		if quote != 0 {
			if ch == quote {
				quote = 0
			}
			buf.WriteRune(ch)
			continue
		}
		if ch == '\'' || ch == '"' {
			quote = ch
			buf.WriteRune(ch)
			continue
		}

		switch {
		case ch == ';':
			segments = append(segments, buf.String())
			buf.Reset()
		case ch == '&' && i+1 < len(runes) && runes[i+1] == '&':
			segments = append(segments, buf.String())
			buf.Reset()
			i++ // skip second &
		case ch == '|' && i+1 < len(runes) && runes[i+1] == '|':
			segments = append(segments, buf.String())
			buf.Reset()
			i++ // skip second |
		default:
			buf.WriteRune(ch)
		}
	}
	if buf.Len() > 0 {
		segments = append(segments, buf.String())
	}
	return segments
}

func anyMatch(patterns []*regexp.Regexp, s string) (*regexp.Regexp, bool) {
	for _, p := range patterns {
		if p.MatchString(s) {
			return p, true
		}
	}
	return nil, false
}

func inputStr(ti map[string]any, key string) string {
	if v, ok := ti[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// --- Output ---

func respondDeny(reason string) {
	out := hookOutput{
		HookSpecificOutput: hookSpecific{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: reason,
		},
	}
	json.NewEncoder(os.Stdout).Encode(out)
	os.Exit(0)
}

func respondAllowWithRewrite(command string) {
	out := hookOutput{
		HookSpecificOutput: hookSpecific{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
			UpdatedInput:       map[string]any{"command": command},
		},
	}
	json.NewEncoder(os.Stdout).Encode(out)
	os.Exit(0)
}

// --- Audit logging ---

func auditLog(toolName string, toolInput map[string]any, decision, reason, mode string) {
	defer func() { recover() }() // never panic

	logDir := filepath.Join(projectDir, ".claude", "workspace", "logs")
	os.MkdirAll(logDir, 0o755)
	logPath := filepath.Join(logDir, "hook-audit.jsonl")

	summary, _ := json.Marshal(toolInput)
	if len(summary) > 256 {
		summary = append(summary[:253], '.', '.', '.')
	}

	entry := map[string]string{
		"ts":       time.Now().UTC().Format(time.RFC3339),
		"sid":      sessionID,
		"tool":     toolName,
		"input":    string(summary),
		"decision": decision,
		"reason":   reason,
		"mode":     mode,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(data)
	f.Write([]byte("\n"))
}

// --- Security checks ---

func isDangerousRm(cmd string) bool {
	n := normalize(cmd)
	if _, ok := anyMatch(dangerousRmPatterns, n); ok {
		return true
	}
	if rmRecursivePattern.MatchString(n) {
		if _, ok := anyMatch(dangerousRmPathPatterns, n); ok {
			return true
		}
	}
	return false
}

func isForkBomb(cmd string) bool {
	n := normalize(cmd)
	_, ok := anyMatch(forkBombPatterns, n)
	return ok
}

// isDangerousGit checks each segment of a chained command independently.
// "git push origin && git checkout main" is allowed (push has no main target).
// "git push origin main && echo done" is blocked (push targets main).
func isDangerousGit(cmd string) string {
	n := normalize(cmd)
	for _, seg := range splitSegments(n) {
		seg = strings.TrimSpace(seg)

		if gitPushMainPattern.MatchString(seg) {
			return "Push to main/master is blocked — use a PR instead"
		}
		if gitPushForcePattern.MatchString(seg) && !gitPushForceLease.MatchString(seg) {
			return "Force push is blocked — use --force-with-lease instead"
		}
		if gitPushDashF.MatchString(seg) {
			return "Force push is blocked — use --force-with-lease instead"
		}
		if gitResetHard.MatchString(seg) {
			return "Dangerous git command blocked"
		}
		if gitCleanFd.MatchString(seg) {
			return "Dangerous git command blocked"
		}
	}
	return ""
}

func isDangerousDisk(cmd string) bool {
	_, ok := anyMatch(dangerousDiskPatterns, normalize(cmd))
	return ok
}

func isNetworkEscapeThreat(cmd string) (bool, string) {
	n := normalize(cmd)
	checks := []struct {
		patterns []*regexp.Regexp
		message  string
	}{
		{cloudMetadataPatterns, "Cloud metadata access blocked (SSRF)"},
		{pipeToShellPatterns, "Piping remote content to shell is blocked"},
		{reverseShellPatterns, "Reverse shell attempt blocked"},
		{containerEscapePatterns, "Container escape attempt blocked"},
		{credentialExfilPatterns, "Credential exfiltration blocked"},
		{cryptoMiningPatterns, "Crypto mining blocked"},
	}
	for _, c := range checks {
		if _, ok := anyMatch(c.patterns, n); ok {
			return true, c.message
		}
	}
	return false, ""
}

func isSensitiveFile(path string) (bool, string) {
	if path == "" {
		return false, ""
	}
	lower := strings.ToLower(path)
	base := filepath.Base(lower)

	// Container mode: allow .ssh/ (only explicit keys are mounted)
	if containerMode && strings.Contains(lower, ".ssh/") {
		return false, ""
	}

	// Allow safe .env variants
	if allowedEnvFiles[base] {
		return false, ""
	}

	// Exact filename matches
	if sensitiveFiles[base] {
		return true, fmt.Sprintf("Access to %s files is prohibited", base)
	}

	// Pattern matches
	for _, sp := range sensitiveFilePatterns {
		if sp.re.MatchString(lower) {
			return true, fmt.Sprintf("Access to %s is prohibited", sp.desc)
		}
	}
	return false, ""
}

func isPathEscape(filePath string) (bool, string) {
	if filePath == "" || projectDir == "" {
		return false, ""
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return true, "Invalid path"
	}
	absProject, err := filepath.Abs(projectDir)
	if err != nil {
		return true, "Invalid project path"
	}

	// Resolve symlinks (only if path exists — EvalSymlinks returns "" for non-existent paths)
	if resolved, err := filepath.EvalSymlinks(absPath); err == nil {
		absPath = resolved
	}
	if resolved, err := filepath.EvalSymlinks(absProject); err == nil {
		absProject = resolved
	}

	if absPath != absProject && !strings.HasPrefix(absPath, absProject+string(os.PathSeparator)) {
		return true, "Path is outside project directory"
	}

	return false, ""
}

func checkBashCommand(cmd string) string {
	if !containerMode {
		if isDangerousRm(cmd) {
			return "Dangerous rm command detected"
		}
		if isForkBomb(cmd) {
			return "Fork bomb detected"
		}
		if isDangerousDisk(cmd) {
			return "Dangerous disk write operation detected"
		}
		if _, ok := anyMatch(dangerousDeletePatterns, normalize(cmd)); ok {
			return "Dangerous delete operation detected"
		}
	}

	if msg := isDangerousGit(cmd); msg != "" {
		return msg
	}

	// .env access — first check if the command matches any env access pattern,
	// then verify each .env reference individually. A command like
	// "cat .env.example .env" must still block the bare ".env".
	for _, p := range envAccessPatterns {
		if p.MatchString(cmd) {
			if hasUnsafeEnvRef(cmd) {
				return "Access to .env files is prohibited"
			}
		}
	}

	// Network / escape threats
	if threat, msg := isNetworkEscapeThreat(cmd); threat {
		return msg
	}

	// Bash access to sensitive paths (cat ~/.aws/credentials, etc.)
	// In container mode, .ssh/ is allowed (only mounted keys are present)
	n := normalize(cmd)
	if _, ok := anyMatch(bashSensitivePathPatterns, n); ok {
		if !(containerMode && strings.Contains(n, ".ssh/")) {
			return "Access to sensitive file via shell is prohibited"
		}
	}

	return ""
}

func checkFileOperation(toolName string, toolInput map[string]any) string {
	filePath := inputStr(toolInput, "file_path")
	if toolName == "Grep" || toolName == "Glob" {
		if p := inputStr(toolInput, "path"); p != "" {
			filePath = p
		}
	}
	if filePath == "" {
		return ""
	}

	if sensitive, reason := isSensitiveFile(filePath); sensitive {
		return reason
	}

	if !containerMode {
		if filepath.IsAbs(filePath) || strings.Contains(filePath, "..") {
			if escape, reason := isPathEscape(filePath); escape {
				return reason
			}
		}
	}
	return ""
}

// --- Settings.json deny list ---

type denyPatterns map[string][]*regexp.Regexp

func loadDenyPatterns() denyPatterns {
	result := denyPatterns{
		"Bash": nil, "Read": nil, "Edit": nil, "Write": nil,
		"Glob": nil, "Grep": nil, "MultiEdit": nil,
	}

	settingsPath := filepath.Join(projectDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return result
	}

	var settings struct {
		Permissions struct {
			Deny []string `json:"deny"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return result
	}

	// Parse "ToolName(pattern)" entries
	entryRe := regexp.MustCompile(`^(\w+)\((.+)\)$`)
	for _, entry := range settings.Permissions.Deny {
		m := entryRe.FindStringSubmatch(entry)
		if m == nil {
			continue
		}
		toolName, globPattern := m[1], m[2]
		if _, ok := result[toolName]; !ok {
			continue
		}

		// Convert glob to regex
		regexStr := regexp.QuoteMeta(globPattern)
		regexStr = strings.ReplaceAll(regexStr, `\*`, `.*`)

		// Trailing " .*" → optional (so "cmd *" matches "cmd" too)
		// QuoteMeta may or may not escape spaces, so check both forms.
		if strings.HasSuffix(regexStr, `\ .*`) {
			regexStr = regexStr[:len(regexStr)-4] + `(\ .*)?`
		} else if strings.HasSuffix(regexStr, ` .*`) {
			regexStr = regexStr[:len(regexStr)-3] + `( .*)?`
		}

		// Flexible whitespace around .*
		regexStr = strings.ReplaceAll(regexStr, `.*\ `, `.*\s*`)
		regexStr = strings.ReplaceAll(regexStr, `\ .*`, `\s*.*`)
		regexStr = strings.ReplaceAll(regexStr, `.* `, `.*\s*`)
		regexStr = strings.ReplaceAll(regexStr, ` .*`, `\s*.*`)

		compiled, err := regexp.Compile("(?i)^" + regexStr)
		if err != nil {
			continue
		}
		result[toolName] = append(result[toolName], compiled)
	}

	return result
}

func checkSettingsDeny(toolName string, toolInput map[string]any, deny denyPatterns) string {
	patterns := deny[toolName]
	if len(patterns) == 0 {
		return ""
	}

	var matchStr string
	if toolName == "Bash" {
		matchStr = inputStr(toolInput, "command")
	} else {
		matchStr = inputStr(toolInput, "file_path")
		if matchStr == "" {
			matchStr = inputStr(toolInput, "path")
		}
	}
	if matchStr == "" {
		return ""
	}

	for _, p := range patterns {
		if p.MatchString(matchStr) {
			return fmt.Sprintf("Blocked by settings.json deny rule: %s", p.String())
		}
	}
	return ""
}

// --- GTK rewrite ---

func tryGtkRewrite(command string) string {
	// Skip if already using gtk, or has pipes/chains
	if strings.Contains(command, "gtk") {
		return ""
	}
	if strings.ContainsAny(command, "|") {
		return ""
	}
	if strings.Contains(command, "&&") || strings.Contains(command, ";") {
		return ""
	}

	// Check if command starts with a rewritable prefix
	matched := false
	for _, prefix := range gtkPrefixes {
		if strings.HasPrefix(command, prefix) {
			matched = true
			break
		}
	}
	if !matched {
		return ""
	}

	// Resolve gtk binary path (use absolute path so it works regardless of CWD)
	gtkBin := filepath.Join(projectDir, fmt.Sprintf(".claude/gtk/gtk-%s-%s",
		strings.ToLower(runtime.GOOS),
		runtime.GOARCH))

	// Check binary exists
	if _, err := os.Stat(gtkBin); err != nil {
		debugLog(fmt.Sprintf("GTK rewrite skipped: binary not found at %s (%v); install gtk or ignore", gtkBin, err))
		return ""
	}

	return gtkBin + " " + command
}

// --- Main ---

func main() {
	var input hookInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		debugLog(fmt.Sprintf("JSON decode error: %v", err))
		os.Exit(1)
	}

	toolName := input.ToolName
	toolInput := input.ToolInput
	mode := "full"
	if containerMode {
		mode = "container"
	}

	debugLog(fmt.Sprintf("Checking tool: %s (mode=%s)", toolName, mode))

	// Load settings.json deny patterns
	deny := loadDenyPatterns()

	// --- Layer 1: Regex security checks ---
	var securityError string

	switch toolName {
	case "Bash":
		securityError = checkBashCommand(inputStr(toolInput, "command"))
	case "Read", "Edit", "Write", "Glob", "Grep", "MultiEdit":
		securityError = checkFileOperation(toolName, toolInput)
	}

	if securityError != "" {
		auditLog(toolName, toolInput, "deny", securityError, mode)
		respondDeny(securityError)
	}

	// --- Layer 2: settings.json deny list ---
	if denyError := checkSettingsDeny(toolName, toolInput, deny); denyError != "" {
		auditLog(toolName, toolInput, "deny", denyError, mode)
		respondDeny(denyError)
	}

	// --- GTK rewrite (Bash only) ---
	if toolName == "Bash" {
		if rewritten := tryGtkRewrite(inputStr(toolInput, "command")); rewritten != "" {
			debugLog(fmt.Sprintf("GTK rewrite: %s", rewritten))
			auditLog(toolName, toolInput, "allow+rewrite", rewritten, mode)
			respondAllowWithRewrite(rewritten)
		}
	}

	// --- Allow ---
	auditLog(toolName, toolInput, "allow", "", mode)
	os.Exit(0)
}
