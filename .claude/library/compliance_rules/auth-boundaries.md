---
description: Authentication and authorization boundary requirements — access control, MFA, least privilege, IDOR prevention
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

# Authentication & Authorization Boundaries

Covers: ISO 27001 A.8.3, A.8.5 | NIS2 Art. 21(2)(i) (human resources security, access control, asset management), 21(2)(j) (MFA, continuous authentication, secured communications) | OWASP ASVS V2, V4

## Authorization — Every Request (ISO A.8.3, ASVS V4.1.1–V4.1.5)

- Authorization checks enforced server-side on a trusted service layer, never client-side only (ASVS V4.1.1)
- Default-deny: access is explicitly granted, not implicitly allowed (ASVS V4.1.3, ISO A.8.3)
- Access controls fail securely — on exception or error, deny access rather than grant it (ASVS V4.1.5)
- User/role attributes used in access control decisions come from server session or signed token, not from request body/headers/cookies that the user can manipulate (ASVS V4.1.2)
- Data queries scoped to authenticated user's permissions — tenant isolation, row-level security (ISO A.8.3)
- API responses return only fields the caller is authorized to see — no over-exposure (ISO A.8.3, A.8.12)

## IDOR Prevention (ASVS V4.2.1)

- Every resource access verifies the requesting user owns or has explicit access to the requested resource ID
- No direct object references without ownership/permission checks
- Both horizontal (other user's data) and vertical (admin-only operations) escalation prevented

## CSRF Protection (ASVS V4.2.2)

- State-changing operations protected by CSRF tokens, SameSite cookies, or Origin header validation
- Anti-CSRF mechanism applied to all authenticated state-changing endpoints

## Authentication — Passwords (ASVS V2.1, V2.4)

- Minimum 12 characters (ASVS V2.1.1); allow up to 64+, cap at 128 (ASVS V2.1.2)
- No truncation before hashing (ASVS V2.1.3)
- All printable Unicode including spaces and emoji permitted (ASVS V2.1.4)
- No composition rules ("must contain uppercase + number + special") (ASVS V2.1.9)
- No periodic forced rotation (ASVS V2.1.10)
- Passwords checked against breached password list (HaveIBeenPwned k-anonymity API or local bloom filter) during registration, login, and change (ASVS V2.1.7)
- Paste and password managers permitted — no `onpaste` prevention, no `autocomplete="off"` on password fields (ASVS V2.1.11)

## Authentication — Password Storage (ASVS V2.4, ISO A.8.5)

- Passwords hashed with adaptive algorithm: Argon2id (preferred), scrypt, bcrypt (cost >= 10), PBKDF2 (>= 100k iterations) — never MD5, SHA1, SHA256 alone (ASVS V2.4.1–V2.4.4)
- Salt at least 32 bits from CSPRNG (ASVS V2.4.2)
- Optional pepper stored separately from DB (KMS, env var) (ASVS V2.4.5)
- Never plaintext, never reversible encryption

## Authentication — Brute Force Protection (ASVS V2.2.1, ISO A.8.5)

- Rate limiting, account lockout, or progressive delays on login endpoints
- No more than 100 failed attempts per hour on a single account (ASVS V2.2.1)
- Secure notifications sent on authentication detail changes (password, email, MFA) (ASVS V2.2.3)

## Authentication — Recovery (ASVS V2.5)

- Reset tokens time-limited, single-use, sent via secure channel (ASVS V2.5.1)
- No password hints or knowledge-based "secret questions" (ASVS V2.5.2)
- Recovery never reveals current password (ASVS V2.5.3)
- No shared or default accounts (root, admin, sa) (ASVS V2.5.4)

## Authentication — Service-to-Service (ASVS V2.10, NIS2 Art. 21(2)(i))

- Intra-service secrets are not static/unchanging credentials (ASVS V2.10.1)
- Service passwords are not defaults (ASVS V2.10.2)
- Service credentials encrypted at rest or in vault, not plain text in config (ASVS V2.10.3)
- No passwords, API keys, or secrets in source code (ASVS V2.10.4)

## Multi-Factor Authentication (NIS2 Art. 21(2)(j), ASVS V4.3.1)

- MFA enforced for administrative/privileged access (ASVS V4.3.1)
- MFA check cannot be bypassed by direct API calls or URL parameters
- MFA status checked server-side per session
- Standard TOTP (RFC 6238) with 30-second window if using TOTP; rate limiting on verification attempts
- Step-up authentication for sensitive operations (password change, payment, data export) — require re-authentication even within active session (ASVS V3.7.1)

## Least Privilege (ISO A.8.3, NIS2 Art. 21(2)(i))

- Code enforces minimum required permissions — no wildcard permissions (`*`)
- RBAC or ABAC implemented; roles are not user-controllable
- Admin functions behind separate authorization layer
- Separation of duties: destructive or sensitive operations require confirmation; self-approval prevented
- Deactivated accounts cannot authenticate; orphaned permissions cleaned up

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
