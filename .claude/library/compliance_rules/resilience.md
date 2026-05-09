---
description: Business continuity and resilience requirements — graceful degradation, retry logic, idempotency, health checks
languages:
- python
- javascript
- typescript
- go
- rust
- java
alwaysApply: false
---

# Business Continuity & Resilience

Covers: NIS2 Art. 21(2)(c) (business continuity, backup management, disaster recovery, crisis management)

## Graceful Degradation (NIS2 Art. 21(2)(c))

- Circuit breakers on external dependency calls — application does not crash entirely on partial failure
- Fallback logic and timeouts on external dependencies
- No infinite retry loops; retries use exponential backoff with jitter and max attempt limits
- Dead letter queues for failed messages in async processing

## State Recovery (NIS2 Art. 21(2)(c))

- Application can recover from unexpected restarts — stateless design or durable state storage
- Database migrations are reversible; backup/restore procedures exist
- Health check endpoints exposed: liveness and readiness probes for orchestrators

## Idempotency (NIS2 Art. 21(2)(c))

- Critical operations (payments, state changes) are idempotent — safe to retry
- Idempotency keys used to detect duplicate requests where applicable
- Operations designed to handle partial completion on retry

## Message Durability (NIS2 Art. 21(2)(c))

- Messages acknowledged only after successful processing
- Persistent queues for critical workflows — no in-memory-only queues for important data
- Queue consumers handle redelivery gracefully

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
