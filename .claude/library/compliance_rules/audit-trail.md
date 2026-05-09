---
description: Audit trail, logging, and monitoring requirements — what to log, format, protection, incident detection
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

# Audit Trail, Logging & Monitoring

Covers: ISO 27001 A.8.15, A.8.16 | NIS2 Art. 21(2)(b), 21(2)(f), Art. 23 | OWASP ASVS V7

## What MUST Be Logged (ISO A.8.15, ASVS V7.1.3, V7.2.1–V7.2.2)

Every security-relevant event must produce an audit record:

- **Authentication**: All login success and failure, logout, MFA challenge/success/failure
- **Authorization failures**: Every denied access attempt with request metadata
- **Access control decisions**: All decisions should be loggable; all failures must be logged (ASVS V7.2.2)
- **State changes on sensitive data**: Create, update, delete operations on user data, configuration, permissions
- **Admin/privileged actions**: All operations performed with elevated privileges, logged separately
- **Input validation failures**: Rejected input that could indicate attack attempts
- **Session lifecycle**: Session creation, destruction, expiry
- **Configuration changes**: Runtime configuration modifications
- **Data access to sensitive resources**: Who accessed what, when (audit access without logging the data itself — ASVS V8.3.5)

## Log Entry Format (ASVS V7.1.4, ISO A.8.15)

Every log entry MUST include:

| Field | Description |
|-------|-------------|
| Timestamp | UTC, ISO 8601 format, NTP-synced source (ASVS V7.3.4) |
| Event type | Classification of the event (auth, access, error, etc.) |
| Severity | Appropriate log level (see below) |
| Actor | User ID, session ID, service account — who performed the action |
| Source | IP address, user agent, originating service |
| Action | What was performed |
| Resource | What was acted upon |
| Outcome | Success or failure |
| Correlation ID | Request ID / trace ID for end-to-end tracing |

Use structured logging format (JSON) for machine parseability.

## What MUST NOT Be Logged (ASVS V7.1.1–V7.1.2)

- Passwords or credential material (session tokens only in hashed form)
- Payment details (credit card numbers, CVV)
- PII beyond what's necessary for the audit record (prefer user ID over email/name)
- Health data, financial data, or other sensitive categories
- Full request/response bodies containing sensitive data

## Log Level Discipline (RFC 5424)

Map application log levels to RFC 5424 syslog severity levels:

| RFC 5424 Severity | Numerical | Application Mapping |
|-------------------|-----------|---------------------|
| Emergency (0) | 0 | System is unusable — total service failure |
| Alert (1) | 1 | Immediate action required — data loss imminent |
| Critical (2) | 2 | Critical conditions — component failure affecting service |
| Error (3) | 3 | Actionable failures requiring attention — request failures, integration errors |
| Warning (4) | 4 | Degradation, unusual conditions, security anomalies — approaching thresholds |
| Notice (5) | 5 | Normal but significant events — configuration changes, auth events |
| Informational (6) | 6 | Key business events, request lifecycle — not high-frequency per-item operations |
| Debug (7) | 7 | Verbose diagnostic data — must be disabled in production or behind level guard |

Rules:
- **ERROR**: Actionable failures requiring attention — not warnings logged as errors
- **WARN**: Degradation, unusual conditions, security-relevant anomalies — not info logged as warnings
- **INFO**: Key business events, request lifecycle — not high-frequency per-item operations
- **DEBUG**: Verbose diagnostic data — must be disabled in production or behind level guard
- Security events (auth success/failure, access denials) should be at NOTICE or WARN, never DEBUG

## Log Protection (ASVS V7.3)

- Log data encoded to prevent log injection — newlines and control characters escaped; user input sanitized before inclusion in log messages (ASVS V7.3.1)
- Logs protected from unauthorized access and modification — restricted permissions, append-only or immutable storage (ASVS V7.3.3)
- Logs written to centralized, tamper-resistant location — not just local stdout

## Log Integrity for Audit (ISO A.8.15, NIS2 Art. 21(2)(b))

- Audit records must not be modifiable or deletable by the application
- Mutable audit logs are a compliance violation
- Log retention meets regulatory requirements (define per national NIS2 transposition)

## Monitoring and Alerting (ISO A.8.16, NIS2 Art. 21(2)(f))

- Application emits metrics for security-relevant events (failed logins, error rates, permission denials)
- Correlation IDs / request IDs propagated through the full call chain
- Rate limiting on sensitive endpoints enables detection of abuse
- Monitoring endpoints are authenticated and do not expose sensitive internals
- Events consumable by SIEM for near-real-time processing

## NIS2 Incident Reporting Support (Art. 23)

Code must produce sufficient telemetry to support the NIS2 reporting timeline:

| Stage | Deadline | What code must enable |
|-------|----------|-----------------------|
| Early warning | 24 hours | Anomaly detection, automated alerts to incident response team |
| Incident notification | 72 hours | Impact assessment data: affected user count, scope, IoC (source IPs, user agents, file hashes) |
| Final report | 1 month after incident notification | Root cause analysis support: sufficient log detail to determine attack vector, timeline reconstruction via correlation IDs. If incident is ongoing, a progress report is due instead; final report due 1 month after resolution |

- Cross-border impact traceability: multi-region deployments log which regions/jurisdictions are affected
- Evidence preservation: logs retained and immutable for the mandated period

## Error Handling for Logging (ASVS V7.4)

- Generic error messages shown to users — no stack traces, SQL errors, or internal paths (ASVS V7.4.1)
- Error may include a unique correlation ID for investigation
- Exception handling at all boundaries (HTTP handlers, service calls, DB calls) (ASVS V7.4.2)
- Global "last resort" error handler catches all unhandled exceptions (ASVS V7.4.3)
- Error paths include logging — catch blocks log the exception with context

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
