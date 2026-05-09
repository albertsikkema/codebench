---
description: Session management and cookie security requirements — token generation, cookie flags, expiry, termination
languages:
- python
- javascript
- typescript
- go
- rust
- java
alwaysApply: false
---

# Session & Cookie Security

Covers: ISO 27001 A.8.5 | NIS2 Art. 21(2)(i) | OWASP ASVS V3

## Session Token Generation (ASVS V3.2)

- New session token generated on user authentication — session regeneration after successful login (ASVS V3.2.1)
- Session tokens have at least 64 bits of entropy — CSPRNG, minimum 16 hex characters (ASVS V3.2.2)
- Session tokens generated using approved cryptographic algorithms (ASVS V3.2.4)
- Session tokens never revealed in URL parameters (ASVS V3.1.1)

## Cookie Flags (ASVS V3.4)

All cookie-based session tokens MUST have:

| Flag | Requirement | Reference |
|------|-------------|-----------|
| `Secure` | Cookie only sent over HTTPS | ASVS V3.4.1 |
| `HttpOnly` | Cookie not accessible to JavaScript | ASVS V3.4.2 |
| `SameSite` | Set to `Lax` or `Strict` | ASVS V3.4.3 |
| `__Host-` prefix | Implies Secure, no Domain, Path=/ | ASVS V3.4.4 |
| `Path` | Set to most precise path when sharing domain | ASVS V3.4.5 |

## Session Termination (ASVS V3.3)

- Logout invalidates the session token server-side, not just the client cookie (ASVS V3.3.1)
- Session timeout enforced: L1 = 30 days idle; L2 = 12h active / 30min idle (ASVS V3.3.2)
- Password change invalidates all sessions except current (ASVS V3.3.3)
- Users can view and log out of active sessions/devices (ASVS V3.3.4)

## Token-Based Sessions (ASVS V3.5)

- Users can revoke OAuth tokens for linked applications (ASVS V3.5.1)
- API authentication uses rotating tokens, not static keys (ASVS V3.5.2)
- JWTs use digital signatures (RS256/ES256, never `none` or HS256 with weak secret), have `exp`, `iat`, `jti` claims (ASVS V3.5.3)

## Session Storage (ASVS V3.2.3, V8.2.2)

- Session tokens stored only in secure cookies or HTML5 sessionStorage — never localStorage
- No passwords, tokens, or PII in localStorage, IndexedDB, or non-httpOnly cookies (ASVS V8.2.2)
- Logout handler clears client-side storage (ASVS V8.2.3)

## Step-Up Authentication (ASVS V3.7.1)

- Sensitive transactions (payments, settings changes, data export) require full valid session or re-authentication
- Re-authentication gate applies even within an active session

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
