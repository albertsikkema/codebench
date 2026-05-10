You are tasked with running a comprehensive automated code analysis on the codebase using the code-index MCP server tools.

## Scope

If the user specifies a file, directory, or module — scope the analysis to that area. Otherwise, analyze the entire project.

## Steps

1. **Project overview:**
   - Run `get_project_summary()` to understand the codebase structure and key symbols.

2. **Run all analysis tools in parallel:**

   Launch these code-index MCP calls concurrently:

   | Tool | Purpose |
   |------|---------|
   | `find_hotspots()` | High-risk files (git churn × complexity) |
   | `find_circular_deps()` | Circular import chains |
   | `analyze_coupling()` | Module coupling metrics (Ca, Ce, instability) |
   | `find_unhandled_errors()` | Error handling gaps |
   | `find_dead_code()` | Potentially unused symbols |
   | `find_duplicates()` | Copy-paste / structural duplication |

   If the user specified a scope, pass the relevant `file`, `file_or_module`, or `language` parameter where supported.

3. **Synthesize findings into a report:**

   Present the results organized by severity:

   ### Critical
   Issues that indicate bugs, security risks, or architectural problems:
   - Circular dependencies (cause import failures, tight coupling)
   - Unhandled errors in critical paths
   - High-instability modules with many dependents (fragile code)

   ### Warnings
   Issues that indicate code quality concerns:
   - Hotspot files (high churn + high complexity — likely to introduce bugs)
   - Highly coupled modules (hard to change independently)
   - Significant code duplication

   ### Info
   Observations for awareness:
   - Dead code candidates (verify before removing — may be used dynamically)
   - Coupling metrics summary
   - Overall codebase health indicators

4. **Actionable recommendations:**
   - Prioritize findings by impact
   - Suggest concrete next steps for the top issues
   - Reference specific files and line numbers

## Output format

Keep the report concise and scannable. Use tables for metrics, bullet points for findings. Skip sections that have no findings — don't include empty headings.

If everything looks clean, say so briefly — don't invent problems.
