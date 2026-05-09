---
description: Cryptography requirements — approved algorithms, key management, TLS, data at rest/transit encryption
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

# Cryptography & Encryption

Covers: ISO 27001 A.8.24 | NIS2 Art. 21(2)(h) | OWASP ASVS V9

## Approved Algorithms (ISO A.8.24, NIS2 Art. 21(2)(h))

### Symmetric Encryption
- **Use**: AES-256-GCM (preferred), AES-256-CBC with HMAC, ChaCha20-Poly1305
- **Never use**: DES, 3DES, RC4, Blowfish, any export-grade cipher

### Asymmetric Encryption / Signatures
- **Use**: RSA-2048+ (prefer 4096), ECDSA P-256+, Ed25519
- **Never use**: RSA < 2048 bits, EC curves < 256 bits

### Hashing (non-password)
- **Use**: SHA-256, SHA-384, SHA-512, BLAKE2, BLAKE3
- **Never use for security**: MD5, SHA-1

### Password Hashing
- **Use**: Argon2id (preferred), scrypt, bcrypt (cost >= 10), PBKDF2-HMAC-SHA256 (>= 100k iterations)
- **Never use**: MD5, SHA1, SHA256 alone, any unsalted hash

### Random Number Generation
- **Use**: CSPRNG only — `crypto/rand` (Go), `secrets` (Python), `crypto.randomBytes` (Node.js), `/dev/urandom`
- **Never use for security**: `math/rand`, `Math.random()`, `random.random()`

## No Custom Cryptography (ISO A.8.24)

- No hand-rolled encryption, hashing, or random number generation
- Use established libraries only: libsodium, OpenSSL, built-in `crypto` modules
- IVs and nonces must be unique per encryption operation — never reused

## Key Management (ISO A.8.24, NIS2 Art. 21(2)(h))

- Keys never hardcoded in source code
- Keys loaded from dedicated KMS, HSM, or secrets manager — not from config files
- Different keys per environment (dev/staging/production)
- Key rotation supported — code must handle key versioning
- Private keys and certificates not committed to repository — `.gitignore` covers `.pem`, `.key`, `.p12`, `.pfx`

## TLS Configuration (ASVS V9.1, V9.2)

### Client-Facing (ASVS V9.1.1–V9.1.3)
- TLS enforced on all client connectivity; no fallback to HTTP for authenticated content
- HSTS header set: `max-age=31536000; includeSubDomains; preload` (ASVS V14.4.5)
- Only TLS 1.2 and TLS 1.3 enabled — SSLv3, TLS 1.0, TLS 1.1 disabled
- Strong cipher suites only; strongest preferred in order

### Server-to-Server (ASVS V9.2.1–V9.2.4)
- Certificate validation enabled — no `verify=False` (Python), no `NODE_TLS_REJECT_UNAUTHORIZED=0` (Node.js), no `InsecureSkipVerify: true` (Go), no `rejectUnauthorized: false` (Node.js)
- Internal service communication uses TLS; no plain HTTP between services crossing trust boundaries
- Mutual TLS or certificate pinning for sensitive integrations
- OCSP stapling properly configured

## Data at Rest (ISO A.8.24, ASVS V8.3.7)

- Sensitive data encrypted in databases using approved algorithms
- Encryption keys stored separately from encrypted data
- Application-level or transparent data encryption where column-level sensitivity warrants it

## Data in Transit (ISO A.8.24, NIS2 Art. 21(2)(h))

- All external communications over TLS
- Internal service-to-service encrypted where crossing trust boundaries
- No sensitive data sent over non-TLS connections (including internal)
- Webhook/notification URLs validated before sending data

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
