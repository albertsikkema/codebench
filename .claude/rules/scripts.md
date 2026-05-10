# Codebase Graph

Generate an interactive HTML visualization of the codebase. Requires the MCP code-index server to have run at least once.

```bash
go run .claude/helpers/codebase-graph/main.go              # file dependency graph
go run .claude/helpers/codebase-graph/main.go -view symbol # symbol call graph
```

See `.claude/library/scripts.md` for full flag reference.
