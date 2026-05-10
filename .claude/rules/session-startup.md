# Session Startup

At the start of every session, before responding to the user's first message:

**Read code index**: Glob for `.claude/index/*.md` then Read each file. These index files are auto-generated and kept up to date by the MCP code-index server — they are regenerated on every file change, branch switch, and server startup. You can trust them as an accurate, current view of the codebase: symbol maps, cross-references, and dependency graphs.
