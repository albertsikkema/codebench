---
description: Secure coding requirements — input validation, output encoding, injection prevention, file handling, API security
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

# Secure Coding

Covers: ISO 27001 A.8.7, A.8.27 (Secure System Architecture and Engineering Principles), A.8.28 | NIS2 Art. 21(2)(a), 21(2)(e) | OWASP ASVS V5, V12, V13

## Input Validation (ASVS V5.1, ISO A.8.28)

- All input validated server-side using positive validation (allowlists, not blocklists) (ASVS V5.1.3)
- Structured data strongly typed with schema validation — JSON Schema, OpenAPI, Zod, Joi, etc. (ASVS V5.1.4)
- Protection against HTTP parameter pollution (ASVS V5.1.1)
- Protection against mass assignment — models use allowlists for bindable fields (`$fillable`, DTOs) (ASVS V5.1.2)
- URL redirects and forwards only allow allowlisted destinations — no open redirect (ASVS V5.1.5)
- File uploads validated: type (magic bytes, not just extension), size limits, filename sanitization (ASVS V12.2.1, V12.1.1)
- Compressed files checked against max uncompressed size — zip bomb prevention (ASVS V12.1.2)

## Output Encoding (ASVS V5.3, ISO A.8.28)

- Context-appropriate encoding at every output point: HTML, JavaScript, URL, CSS, HTTP header, SMTP (ASVS V5.3.1)
- Template engine auto-escaping enabled; manual escaping for JS/CSS/URL contexts (ASVS V5.3.3)
- Content-Type headers set with safe charset on every response (ASVS V14.4.1)

## Injection Prevention (ASVS V5.3.4–V5.3.10, ISO A.8.28)

| Attack | Prevention | Reference |
|--------|-----------|-----------|
| SQL injection | Parameterized queries or ORM; no string concatenation in queries | ASVS V5.3.4 |
| OS command injection | Parameterized command execution; no `shell=True` with user input | ASVS V5.3.8 |
| XSS | Context-aware output encoding; CSP headers | ASVS V5.3.3, V14.4.3 |
| LDAP injection | Parameterized LDAP filters or proper escaping | ASVS V5.3.7 |
| JSON injection | `JSON.parse()` not `eval()`; no user input interpolated into JSON strings | ASVS V5.3.6 |
| XPath/XML injection | Parameterized queries; safe XML builders | ASVS V5.3.10 |
| Template injection | Auto-escaping; user input never used as template source | ASVS V5.2.5 |
| SMTP injection | User input sanitized before passing to mail systems | ASVS V5.2.3 |
| LFI/RFI | File paths not constructed from user input; allowlist validation | ASVS V5.3.9, V12.3.1–V12.3.3 |

## Dynamic Code Execution (ASVS V5.2.4, ISO A.8.7)

- No `eval()`, `exec()`, `new Function()`, `setTimeout(string)` on user-controlled input
- No deserialization of untrusted data without safeguards — no `pickle.loads()`, `ObjectInputStream`, `yaml.load()` without SafeLoader (ASVS V5.5.3)
- XML parsers configured with external entity resolution disabled (ASVS V5.5.2)
- No execution of dynamically downloaded code without integrity checks

## Sanitization (ASVS V5.2)

- Untrusted HTML from WYSIWYG editors sanitized with DOMPurify, bleach, or equivalent (ASVS V5.2.1)
- User-supplied SVG strips `<script>`, `onload=`, `<foreignObject>` (ASVS V5.2.7)
- User-supplied Markdown/CSS/BBCode sanitized or sandboxed (ASVS V5.2.8)

## SSRF Prevention (ASVS V5.2.6, V12.6.1)

- Outbound HTTP calls validate URLs against allowlist for protocols, domains, paths, ports
- Block `file://`, `gopher://`, internal IPs (127.0.0.1, 169.254.x.x, 10.x.x.x, 172.16-31.x.x, 192.168.x.x)
- Web/application server configured with URL allowlist for outbound resource loading

## File Handling (ASVS V12.3–V12.5, ISO A.8.28)

- User-submitted filenames sanitized or replaced with generated names — no path traversal (ASVS V12.3.1)
- Files from untrusted sources stored outside web root with limited permissions (ASVS V12.4.1)
- Served uploads use `Content-Type: application/octet-stream` or `Content-Disposition: attachment` — never execute as HTML/JS (ASVS V12.5.2)
- Web server blocks serving `.bak`, `.swp`, `.old`, `.tmp`, `.sql`, `.log` files (ASVS V12.5.1)

## API Security (ASVS V13)

- Consistent character encoding (UTF-8) across all layers (ASVS V13.1.1)
- No secrets in API URLs (ASVS V13.1.3)
- Authorization at both URI and resource level (ASVS V13.1.4)
- Only valid HTTP methods accepted per route (ASVS V13.2.1)
- JSON schema validation on input before processing (ASVS V13.2.2)
- REST services with cookies protected from CSRF (ASVS V13.2.3)
- GraphQL: query depth/complexity limiting; auth in resolvers not schema directives (ASVS V13.4.1–V13.4.2)

## HTTP Security Headers (ASVS V14.4)

| Header | Value | Reference |
|--------|-------|-----------|
| `Content-Security-Policy` | No `unsafe-inline`, `unsafe-eval` unless absolutely necessary | ASVS V14.4.3 |
| `X-Content-Type-Options` | `nosniff` | ASVS V14.4.4 |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains; preload` | ASVS V14.4.5 |
| `Referrer-Policy` | `strict-origin-when-cross-origin` or stricter | ASVS V14.4.6 |
| `X-Frame-Options` / CSP frame-ancestors | `DENY` or `'self'` | ASVS V14.4.7 |
| CORS `Access-Control-Allow-Origin` | Strict allowlist; no wildcard `*` on authenticated endpoints | ASVS V14.5.3 |

## Memory Safety (ASVS V5.4 — C/C++/Rust)

- Memory-safe string operations; bounds checking on arrays (ASVS V5.4.1)
- No user-controlled format strings in `printf`, `String.format`, logging (ASVS V5.4.2)
- Range checks on arithmetic with user-supplied integers (ASVS V5.4.3)

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
