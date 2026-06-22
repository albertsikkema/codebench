# Production Audit Report: {project_name}

**Date**: {date}
**Repository**: {repo_path}
**Language(s)**: {languages}
**Framework(s)**: {frameworks}
**Deployment**: {deployment_platform}
**Production URL(s)**: {production_urls or "not identified"}

---

## 1. Overall Assessment

**Risk Rating**: {CRITICAL | HIGH | MEDIUM | LOW}

Rollup rule: any CRITICAL finding = CRITICAL overall. Otherwise, any
HIGH = HIGH overall. And so on.

{2-3 sentence summary of the most significant findings}

**Verdict**: {Production-ready | Not production-ready}. {1 sentence why}.

### Severity Count by Category

| Category | Critical | High | Medium | Low |
|----------|----------|------|--------|-----|
| Architecture and Structure (ARC) | | | | |
| Security (SEC) | | | | |
| Data Governance / GDPR (DG) | | | | |
| Observability and Logging (OBS) | | | | |
| Resilience and Error Handling (RES) | | | | |
| Code Quality (CQ) | | | | |
| Accessibility (A11Y) | | | | |
| Dependencies (DEP) | | | | |
| Infrastructure and Deployment (INF) | | | | |
| Documentation (DOC) | | | | |
| **Total** | | | | |

Only include categories that have findings. Remove empty rows.

---

## 2. Findings by Category

Organize findings by category. Within each category, order by severity
(CRITICAL first, then HIGH, MEDIUM, LOW).

Each finding format:

```
**{CAT}-{S}{N}. {One-line title}** -- {SEVERITY}
- {file:line evidence}
- {What was observed}
```

Where:
- {CAT} = category prefix (ARC, SEC, DG, OBS, RES, CQ, A11Y, DEP, INF, DOC)
- {S} = severity initial (C, H, M, L)
- {N} = sequential number within that category+severity

Example: `SEC-C1`, `SEC-C2`, `SEC-H1`, `RES-M3`, `DOC-L2`

### Architecture and Structure (ARC)

{findings}

### Security (SEC)

{findings}

### Data Governance / GDPR (DG)

{findings -- include PII data flow summary at end of section}

### Observability and Logging (OBS)

{findings}

### Resilience and Error Handling (RES)

{findings}

### Code Quality (CQ)

{findings}

### Accessibility (A11Y)

{findings -- label WCAG conformance level (A or AA) for each}

### Dependencies (DEP)

{SBOM table first, then findings}

### Infrastructure and Deployment (INF)

{findings}

### Documentation (DOC)

{findings -- include documentation completeness scorecard}

---

## 3. Structural Analysis

### Codebase Profile

| Metric | Value |
|--------|-------|
| Total files (source) | |
| Total symbols | |
| Languages | |
| Total LOC | |
| Contributors | |
| Commits | |
| Repo age | {days} ({created date}; activity summary) |
| Test files | |
| Test coverage | |

### Hotspot Files

| Risk | File | Changes | Complexity |
|------|------|---------|------------|
| | | | |

{interpretation -- e.g. concentration risk, expected defect rate vs baseline}

### Bus Factor

{bus factor per file or area, single contributor warnings}

### Top Bloated Functions (5+ signals)

| Function | File:Line | LOC | Vars | Calls | Branches | Nesting |
|----------|-----------|-----|------|-------|----------|---------|
| | | | | | | |

{N total bloated functions out of M analyzed}

### Top Cognitive Complexity

| Function | Complexity | Threshold |
|----------|------------|-----------|
| | | |

{N functions exceed threshold 10. M exceed threshold 15.}

### Duplicates

{count of duplicate pairs, then primary clusters grouped by pattern:}
- {N identical X functions}
- {N identical Y handlers}
- {consolidation opportunity summary}

### Nesting

{N functions with nesting depth >3 (worst: depth N)}

### Architecture Metrics

- Circular dependencies: {count or "none"}
- Coupling: {summary -- Ca/Ce metrics or "zero coupling" with explanation}
- Module system: {ES modules / CommonJS / script tags / none}
- Temporal coupling: {summary or "none detected" with explanation}
- Dead code: {summary}

---

## 4. Production Verification Results

If production URL(s) were checked, record results here.
If not checked, state why (no URL known, write risk, etc.).

### Security Headers

Verified via `curl -sI {production_url}`

| Header | Present | Value |
|--------|---------|-------|
| `Strict-Transport-Security` | | |
| `Content-Security-Policy` | | |
| `X-Frame-Options` | | |
| `X-Content-Type-Options` | | |
| `Referrer-Policy` | | |
| `Permissions-Policy` | | |

### Webroot Exposure

| Path | Status | Should be public? |
|------|--------|-------------------|
| | | |

### Stale / Duplicate Deployments

{any orphaned resources, old deployments still serving traffic}

### Discrepancies Between Repo and Production

{findings where production behavior differs from what the repo config
suggests -- e.g., headers added by platform defaults}

---

## 5. Manual Verification Checklist

Items requiring human execution. Mark each as:
- `[x]` **VERIFIED** -- checked, with result
- `[ ]` **SKIPPED** -- not checked, with reason
- `[x]` **FAIL** -- checked, found a problem

### Runtime Verification
- [ ] App starts and responds
- [ ] Golden path: login, core feature, save
- [ ] Auth bypass: access without session
- [ ] IDOR: access another user's resource by manipulating IDs
- [ ] XSS: inject script via user input / file import
- [ ] Error responses: no stack traces in error output
- [ ] Security headers: curl -I (may be done in Phase 4)
- [ ] CORS: credentialed endpoints not using wildcard origin
- [ ] Offline behavior: disconnect during editing

### Manual Code Review
- [ ] Logic correctness: spot-check highest-complexity function
- [ ] Verify escape/sanitization functions are correctly implemented
- [ ] Audit innerHTML/dangerouslySetInnerHTML assignments
- [ ] Verify data files contain fictional data (not real PII)
- [ ] Review API authentication: are dev bypasses trusted in production?
- [ ] Verify data storage is scoped per user/tenant

---

## 6. Appendix: Positive Findings

{Things that are clean, correctly implemented, or well-designed}

---

## 7. Metadata

- **Audit date**: {date}
- **Phases completed**: {list which phases ran}
- **Tools used**: code-index MCP, audit agents, curl, WebFetch
- **Limitations**: {anything not checked and why}
