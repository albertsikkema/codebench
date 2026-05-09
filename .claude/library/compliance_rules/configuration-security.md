---
description: Configuration security requirements — secure defaults, environment separation, secrets management, hardening
languages:
- python
- javascript
- typescript
- go
- rust
- java
alwaysApply: false
---

# Configuration Security

Covers: ISO 27001 A.8.9, A.8.31, A.8.33 | OWASP ASVS V14.1, V14.2, V14.3

## Secure Defaults (ISO A.8.9, ASVS V14.3.2)

- No default credentials or default configuration values left in code
- Debug modes disabled in production — no `DEBUG=true`, `app.debug=True`, `NODE_ENV=development` in production config (ASVS V14.3.2)
- Features disabled by default; opt-in for dangerous operations (ISO A.8.9)
- No `Server`, `X-Powered-By`, `X-AspNet-Version` headers exposing technology versions (ASVS V14.3.3)

## Secrets Management (ISO A.8.9, ASVS V2.10.3–V2.10.4)

- Secrets loaded from secrets management systems (Vault, AWS Secrets Manager, Azure Key Vault, environment variables)
- Never committed to source control — `.gitignore` covers `.env`, `.pem`, `.key`, `.p12`, `.pfx`, `credentials.json`
- No hardcoded passwords, API keys, tokens, or connection strings in source code
- Pre-commit hooks prevent accidental secret commits (git-secrets, detect-secrets, or equivalent)

## Externalized Configuration (ISO A.8.9)

- Security-relevant configuration explicitly set, not left to framework defaults:
  - TLS versions and cipher suites
  - CORS policies
  - CSP headers
  - Session timeout values
  - Rate limiting thresholds
- Configuration externalized via environment variables or config files — not hardcoded
- Database connection strings use TLS/encryption
- Configuration changes are auditable (version-controlled config)

## Environment Separation (ISO A.8.31 — Separation of Development, Test and Production Environments, A.8.33)

- No environment-specific security bypasses: no `if (env === 'production') skip_auth()` patterns (ISO A.8.31)
- Configuration cleanly separates environments without security degradation
- No hardcoded production URLs/credentials in code — loaded from environment
- Test fixtures/seeds do not contain real production data (ISO A.8.33)
- Test configurations do not reference production databases (ISO A.8.33)

## Infrastructure Hardening (ISO A.8.9, ASVS V14.1)

- Dockerfiles use hardened base images and disable unnecessary services/ports
- No unused services, ports, or endpoints in production
- Compiler flags enable buffer overflow protections where applicable: `-fstack-protector-strong`, `-D_FORTIFY_SOURCE=2` (ASVS V14.1.2)
- Directory browsing disabled; no `.git`, `.env`, `web.config` accessible from web (ASVS V4.3.2)

## Build and Deploy (ASVS V14.1)

- CI/CD pipeline exists; no manual deployment steps (ASVS V14.1.1)
- Application and dependencies re-deployable via automated scripts (ASVS V14.1.4)
- Production config follows framework security hardening guide (ASVS V14.1.3)

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
