# Codebase Graph

Generate an interactive HTML visualization of the codebase. Requires the MCP code-index server to have run at least once.

```bash
go run .claude/helpers/codebase-graph/main.go   # generates single HTML with file/symbol/health views
```

Output: `.claude/helpers/codebase-graph/codebase-graph.html` (switchable via dropdown).

See `.claude/library/scripts.md` for full flag reference.
