---
description: Data lifecycle requirements — data protection, masking, leakage prevention, retention, deletion, classification
languages:
- python
- javascript
- typescript
- go
- rust
- java
alwaysApply: false
---

# Data Lifecycle & Protection

Covers: ISO 27001 A.8.10, A.8.11, A.8.12 | OWASP ASVS V8

## Data Classification (ISO A.8.12, ASVS V8.3.4)

- All sensitive data identified with a handling policy — sensitive fields marked in schema/models
- Classification levels applied: public, internal, confidential, restricted
- Processing rules follow classification — higher classification = stricter controls

## Data Leakage Prevention (ISO A.8.12)

- No sensitive data (PII, credentials, financial data) written to log files (ISO A.8.12, ASVS V7.1.1)
- API responses return only necessary fields — no over-exposure in list endpoints (ISO A.8.12)
- No sensitive data in URL query parameters — use POST body or headers (ASVS V8.3.1)
- Database queries return only needed columns/rows, not `SELECT *` on sensitive tables
- Outbound HTTP calls do not inadvertently send sensitive data to third parties
- No sensitive data in error messages exposed to users
- Email/notification templates do not include raw sensitive data

## Client-Side Data Protection (ASVS V8.2)

- Anti-caching headers on sensitive data responses: `Cache-Control: no-store`, `Pragma: no-cache` (ASVS V8.2.1)
- No passwords, tokens, or PII in localStorage, sessionStorage, IndexedDB, or non-httpOnly cookies (ASVS V8.2.2)
- Authenticated data cleared from client storage on logout (ASVS V8.2.3)

## Server-Side Data Protection (ASVS V8.1)

- Application-level caches (Redis, Memcached) exclude or encrypt sensitive fields (ASVS V8.1.1)
- Cached/temporary copies of sensitive data have TTL and are encrypted or access-controlled (ASVS V8.1.2)
- Forms do not include unnecessary hidden fields with sensitive data (ASVS V8.1.3)
- Rate limiting or anomaly detection on sensitive endpoints (ASVS V8.1.4)

## Data Masking (ISO A.8.11)

- Sensitive data displayed in UIs is masked (e.g., last 4 digits of credit card, redacted email)
- PII masking in non-production environments — test and staging use anonymized data
- Log output redacts/masks sensitive fields before writing

## Data Retention & Deletion (ISO A.8.10, ASVS V8.3.8)

- Personal data stored with defined retention limits — TTL or auto-deletion schedule
- Sensitive personal information subject to data retention classification (ASVS V8.3.8)
- Deletion is real, not just soft-delete, for data that must be purged (ISO A.8.10)
- Database migrations do not destroy data without backup provisions (ISO A.8.13)

## Data Portability & Erasure (ASVS V8.3.2–V8.3.3)

- Users can export their data on demand — data export endpoint exists (ASVS V8.3.2)
- Users can request deletion — data deletion endpoint exists (ASVS V8.3.2)
- Consent mechanism exists in registration/data collection flows (ASVS V8.3.3)

## Sensitive Data in Memory (ASVS V8.3.6)

- Sensitive buffers zeroed when no longer needed — `SecureString`, `Arrays.fill()`, `memset_s()`
- Applies to passwords, encryption keys, tokens held in memory

## Data Encryption (ASVS V8.3.7)

- Sensitive/private information encrypted using approved algorithms (AES-256, ChaCha20)
- See `cryptography.md` for full algorithm requirements

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
