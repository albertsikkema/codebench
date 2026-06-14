# Code Audit Report: {project_name}

**Date**: {date}
**Auditor**: {auditor}
**Repository**: {repo_path}
**Language(s)**: {languages}
**Framework(s)**: {frameworks}

---

## Executive Summary

**Overall Risk Rating**: {CRITICAL | HIGH | MEDIUM | LOW}

{2-3 sentence overall assessment}

| Category | Critical | High | Medium | Low |
|----------|----------|------|--------|-----|
| Security | | | | |
| Data Governance | | | | |
| Code Quality | | | | |
| Observability | | | | |
| Resilience | | | | |
| Dependencies | | | | |
| Documentation | | | | |
| **Total** | | | | |

---

## Critical Findings (MUST FIX)

{Findings that represent active security vulnerabilities or compliance violations}

---

## High Priority Findings

{Findings that represent significant risk or quality issues}

---

## Medium Priority Findings

{Findings that should be addressed but do not represent immediate risk}

---

## Low Priority / Informational

{Observations and recommendations for improvement}

---

## Structural Analysis

{Output from code-index MCP tools: hotspots, complexity, duplication, dead code, architecture}

---

## Detailed Agent Reports

<details>
<summary>Security Audit</summary>

{Full security agent output}

</details>

<details>
<summary>Data Governance Audit</summary>

{Full data governance agent output}

</details>

<details>
<summary>Observability Audit</summary>

{Full observability agent output}

</details>

<details>
<summary>Resilience Audit</summary>

{Full resilience agent output}

</details>

<details>
<summary>Dependencies Audit</summary>

{Full dependencies agent output}

</details>

<details>
<summary>Documentation Audit</summary>

{Full documentation agent output}

</details>

---

## Manual Verification Checklist

Phase 8 (runtime) and Phase 10 (manual review) items requiring human execution:

### Runtime Verification (requires app running)
- [ ] App starts and responds to health check
- [ ] Golden path walkthrough (signup, core feature, data CRUD)
- [ ] Auth bypass test (access endpoint without session)
- [ ] IDOR test (access another user resource)
- [ ] XSS test in input fields
- [ ] Error response content (no stack traces in 500 responses)
- [ ] Security headers present (curl -I)
- [ ] CORS not wildcard on credentialed endpoints

### Manual Code Review
- [ ] Logic correctness spot-checks on critical paths
- [ ] Hallucinated API verification (do imported packages/methods exist?)
- [ ] Test quality assessment (meaningful assertions vs mock wiring?)
- [ ] Name/comment quality (misleading names, restating-the-obvious?)
- [ ] Over-abstraction review (single-implementation interfaces, deep hierarchies?)

---

## References

- Research: .claude/memories/2026-06-14-vibe-coding-audit-research.md
- Security rules: .claude/library/security_rules/
- Compliance rules: .claude/library/compliance_rules/
