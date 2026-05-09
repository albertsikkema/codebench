---
description: Secure development lifecycle requirements — code review, security testing, change management, vulnerability disclosure
languages:
- python
- javascript
- typescript
- go
- rust
- java
- c
- cpp
alwaysApply: false
---

# Secure Development Lifecycle

Covers: ISO 27001 A.8.4, A.8.25 (Secure Development Lifecycle), A.8.26, A.8.29 (Security Testing in Development and Acceptance) | NIS2 Art. 21(2)(e)

## Code Review Process (ISO A.8.25, A.8.4, NIS2 Art. 21(2)(e))

- Peer review (code review) mandatory before merge — no direct commits to main/protected branches
- Security-sensitive changes require additional reviewer with security knowledge
- Branch protection rules enforce review requirements
- No commented-out code with secrets; no unresolved `TODO: fix security` or `FIXME: vulnerable` left in code

## Security Testing (ISO A.8.29, NIS2 Art. 21(2)(e))

- Automated security testing in CI pipeline: SAST (static analysis), SCA (dependency scanning), secret scanning, container scanning
- Security-relevant code paths have test coverage: auth, authz, input validation, crypto operations
- Penetration test findings have corresponding automated regression checks
- SAST findings reviewed; no suppressed findings without documented justification

## Change Management (ISO A.8.25, NIS2 Art. 21(2)(e))

- Commits reference tickets/issues for traceability
- PRs require review before merge
- Changes go through defined approval process before production deployment
- Configuration changes tracked in version control

## Application Security Requirements (ISO A.8.26)

- Authentication and authorization requirements implemented as specified
- Input validation and output encoding mechanisms in place
- Data protection requirements (encryption, masking) met
- Transactional integrity controls exist (CSRF tokens, idempotency keys)
- Third-party integrations meet security requirements (TLS, authentication, data minimization)
- API contracts enforce security constraints (required auth headers, rate limits, input schemas)

## Vulnerability Handling and Disclosure (NIS2 Art. 21(2)(e))

- `SECURITY.md` or `security.txt` present with contact information for vulnerability reports
- Coordinated vulnerability disclosure process in place
- Defined process for triaging and patching reported vulnerabilities
- Security patches applied within defined SLAs

## Repository Access Control (ISO A.8.4)

- Read and write access to source code managed with least privilege
- Development tools and software libraries access controlled
- Access to production secrets/config limited to authorized personnel

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
