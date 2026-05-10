# MCP Server Usage Guide

This project has 3 MCP servers configured (`.mcp.json`): `context7`, `code-index`, and `playwright`. Use them as described below.

---

## Code Index Server (`code-index`)

Tree-sitter AST-based code indexer. Auto-indexes the project on startup, watches for file changes. Supports Python, JS/TS, Go, C/C++, Rust.

You MUST use code-index MCP tools FIRST when exploring or navigating the codebase. This is not optional.

| Task | Do NOT use | ALWAYS use instead |
|------|-----------|-------------------|
| Find a function/class/symbol definition | Grep, Glob | `find_symbol(name)` |
| Find all callers / call sites | Grep | `find_usage(name)` |
| Understand a file's structure | Read (full file) | `get_file_outline(file_path)` |
| Get project overview | Glob + Read | `get_project_summary(compact?)` |
| Trace data flow through code | Grep + Read | `trace_data_flow(fn, var, direction)` |
| Search by description or fuzzy name | Grep | `search_symbols(query, kind?, language?, limit?)` |
| Find callers/callees tree | Grep | `get_call_graph(name)` |
| Find subclasses/implementations | Grep | `find_implementations(name)` |
| Detect circular imports | Manual analysis | `find_circular_deps()` |
| Measure module coupling | Manual analysis | `analyze_coupling(file_or_module?)` |
| Search for AST patterns | Grep | `ast_query(pattern, language, path?, limit?)` |
| Find error handling issues | Manual review | `find_unhandled_errors(language?, file?, limit?)` |
| Find risky code hotspots | git log + manual | `find_hotspots(since?, min_changes?, limit?)` |

STOP and use code-index instead if you catch yourself reaching for Grep, Glob, or Read to find or understand code structure.

### When to use tools in a typical workflow

1. **"What does this project look like?"** → `get_project_summary(compact?)` — **the single most important call.** Returns a comprehensive markdown overview of the entire codebase: most-used symbols ranked by reference count, every file with its functions/classes/signatures and call relationships ("used by" cross-references), detected API endpoints, and a dependency graph showing which files are most depended upon. Use `compact=true` to get only stats, top references, and dependency graph without per-file symbol listings. Call this first when starting any task to get the full picture without needing to search.
2. **"Where is X defined?"** → `find_symbol(name, kind?)` — drill into a specific symbol, get its qualified ID. Filter by `kind`: function, method, class, struct, interface, type_alias, enum, model, component.
3. **"Show me the source of X"** → `get_symbol_source(id)` — retrieves exact source by qualified ID. Use `verify=true` to detect if source has changed since indexing. Use `context_lines` (0-50) for surrounding context.
4. **"What calls X?" / "Where is X used?"** → `find_usage(name)` — returns all call sites with caller context and line numbers, not just text matches. Essential before changing a symbol.
5. **"What implements/extends X?"** → `find_implementations(name)` — finds all subclasses, interface implementations, trait impls. Essential before refactoring a base class.
6. **"Find something by name or description"** → `search_symbols(query, kind?, language?, limit?)` — three-tier search: exact name, fuzzy match (typo-tolerant), TF-IDF semantic search across names+signatures+docstrings. Filter by `kind` and/or `language` to narrow results. Use when you don't know the exact name.
7. **"How does X relate to Y?"** → `get_call_graph(name, depth)` — shows callers + callees tree, depth 1-5 (default 2).
8. **"What does this file contain?"** → `get_file_outline(file_path)` — structured list of all symbols with kinds, signatures, and line ranges.
9. **"What does this file depend on?"** → `get_file_dependencies(file_path)` — imports + reverse imports (who imports this file).
10. **"Show class hierarchy"** → `get_type_hierarchy(name, depth?)` — ancestors + descendants tree, depth 1-5 (default 5).
11. **"What uses this decorator?"** → `find_by_decorator(decorator)` — e.g., `"dataclass"`, `"mcp.tool"`, `"pytest.fixture"`.
12. **"What's related to X?"** → `find_related_symbols(name, limit?)` — co-occurrence analysis across call graphs.
13. **"What might be unused?"** → `find_dead_code(language?, confidence?, exclude_frameworks?, limit?)` — symbols with zero cross-file references. Filter by `confidence` (high/medium/low), use `exclude_frameworks=false` to include low-confidence framework symbols.
14. **"Find copy-paste code"** → `find_duplicates(threshold?, language?, min_lines?, limit?)` — structural hash + normalized source comparison. `min_lines` sets minimum function size (default 5).
15. **"Where does this data come from / go?"** → `trace_data_flow(start_function, variable?, direction?, depth?)` — traces data forward (to callees) or backward (to callers) through function call chains. Supports Python, JavaScript/TypeScript, and Go. Uses on-demand intraprocedural analysis with enriched call edges.
16. **Regenerate markdown indexes** → `generate_index(directory?, output?, language?)` or `reindex()` for a full rebuild.
17. **"Are there circular imports?"** → `find_circular_deps()` — DFS cycle detection on the resolved import graph. Returns each cycle as a file chain.
18. **"How coupled is this module?"** → `analyze_coupling(file_or_module?)` — Ca (afferent), Ce (efferent), instability metrics per file. Optional path prefix filter.
19. **"Find AST patterns"** → `ast_query(pattern, language, path?, limit?)` — run tree-sitter S-expression queries across the codebase. Supports all 5 languages. `limit` caps matches (1-500, default 100).
20. **"Any error handling issues?"** → `find_unhandled_errors(language?, file?, limit?)` — detects bare/broad/empty except (Python), unchecked err (Go), await without try/catch (JS/TS), .unwrap() in non-test code (Rust). `limit` caps results (1-500, default 100).
21. **"What are the riskiest files?"** → `find_hotspots(since?, min_changes?, limit?)` — combines git change frequency with code complexity to identify high-risk areas.

### When to use `trace_data_flow` instead of `find_symbol` / `get_call_graph`

`trace_data_flow` is the right tool whenever the question is about **data** rather than **code structure**. Use it when:

- **Tracing user input**: "How does the request body reach the database?" → `trace_data_flow("handle_request", "request", "forward")`
- **Finding data sources**: "Where does this variable's value come from?" → `trace_data_flow("render_page", "user", "backward")`
- **Security review**: "Can user input reach this SQL query?" → forward trace from the HTTP handler
- **Understanding data transformations**: "What happens to this config value?" → forward trace from where it's loaded
- **Debugging**: "Why is this field None/wrong?" → backward trace from where it's used

**Do NOT use `find_symbol` or `get_call_graph` for these questions.** Those tools show structural relationships (who calls whom) but not which specific variables/parameters carry the data between functions. `trace_data_flow` follows the actual data through argument passing, assignments, and return values across function boundaries.

**Rule of thumb**: If the question mentions a *variable*, *parameter*, *field*, *input*, *value*, or *data* flowing/passing/reaching somewhere → use `trace_data_flow`. If the question is about *which functions exist* or *who calls whom* → use `find_symbol` / `get_call_graph`.

### When to use the code analysis tools

These tools perform **automated code analysis** — use them proactively during code reviews, architecture assessments, and quality audits:

- **Architecture review**: `find_circular_deps()` + `analyze_coupling()` → detect structural problems
- **Code review**: `find_unhandled_errors(file="changed_file.py")` → catch error handling gaps in changed files
- **Quality audit**: `find_hotspots()` → identify high-risk areas that need attention
- **Custom linting**: `ast_query(pattern, language)` → find any structural pattern in the codebase
- **Security review**: `find_unhandled_errors()` + `trace_data_flow()` → find both input handling and error handling issues

Use `find_unhandled_errors` and `find_hotspots` PROACTIVELY during `/review` and `/pr-review` workflows.

### When to still use Grep/Glob

ONLY fall back to Grep/Glob/Read when code-index tools genuinely cannot help:
- Searching for string literals, comments, or configuration values (not symbol definitions)
- Finding files by name pattern (not by code content)
- Searching for patterns that aren't symbol definitions (e.g., TODO comments, hardcoded URLs)
- Reading non-code files (config, markdown, YAML, JSON)

NEVER use Grep to find function/class definitions — use `find_symbol(name)`.
NEVER use Grep to find call sites — use `find_usage(name)`.
NEVER use Read on an entire file to understand its structure — use `get_file_outline(file_path)`.

### Skipped directories

The server skips `.claude/`, `node_modules/`, `venv/`, `.venv/`, `examples/`, `dist/`, `build/`, `.git/`, and other non-source directories by default. Use `generate_index(directory="path")` to explicitly index a skipped directory.

---

## Context7 (`context7`)

Up-to-date documentation and code examples for any library/framework.

### When to use

- Before using an unfamiliar library API — get current docs instead of relying on training data
- When you need code examples for a specific library feature
- When checking version-specific behavior or breaking changes

### Workflow

1. `resolve-library-id(libraryName, query)` → get the Context7 library ID
2. `query-docs(libraryId, query)` → get documentation and code examples

Always resolve the library ID first. Max 3 calls per question.

---

## Playwright (`playwright`)

Headless browser automation for testing and visual verification.

### When to use

- Verifying web UI renders correctly (use with `/visual-verify` skill)
- Testing user flows in a browser
- Taking screenshots for visual comparison
- Filling forms, clicking buttons, checking accessibility snapshots

### Key tools

- `browser_navigate(url)` → open a page
- `browser_snapshot()` → get accessibility tree (preferred over screenshots for actions)
- `browser_take_screenshot()` → visual capture
- `browser_click(ref)` / `browser_type(ref, text)` → interact with elements
- `browser_fill_form(fields)` → fill multiple form fields at once
- `browser_select_option(ref, values)` → dropdown selection
- `browser_tabs(action, index?)` → list/create/close/select tabs
- `browser_wait_for(text?, textGone?, time?)` → wait for text to appear/disappear
- `browser_evaluate(function, ref?)` → run JS on page or element
- `browser_console_messages()` / `browser_network_requests()` → debug

### Workflow

Navigate → snapshot → interact → verify. Always use `browser_snapshot` to get element refs before clicking/typing.

Playwright is the **default browser tool** — use it for all UI testing, form filling, user flow automation, and visual verification.

### Containerized runtime — reaching the host's localhost

Claude runs inside a `claude-safe` container, so `http://localhost:PORT` and `http://127.0.0.1:PORT` resolve to the container itself, not the user's host machine. A dev server running on the host is **not** reachable that way.

When the user gives you a `localhost` / `127.0.0.1` URL for a service running on their host, swap the host part for `host.docker.internal` before navigating:

- `http://localhost:5173` → `http://host.docker.internal:5173`
- `http://127.0.0.1:5500/test.html` → `http://host.docker.internal:5500/test.html`

If `host.docker.internal` itself fails to resolve (Linux without Docker Desktop and without `--add-host=host.docker.internal:host-gateway`), tell the user — the container has no path to the host's loopback and they need to either pass that flag, run the dev server inside the container, or expose it on a routable interface. Don't silently fall back to `localhost`.

External URLs (anything not on the host's loopback) work normally, subject to the firewall allowlist when `claude-safe` is run without `--no-firewall`.
