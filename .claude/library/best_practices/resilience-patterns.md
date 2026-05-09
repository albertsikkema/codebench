# Resilience Patterns

## Principle

Every outbound call will eventually fail. Design for failure by applying timeouts, retries with backoff, circuit breakers, graceful degradation, and backpressure. These patterns are independent and composable — use the combination appropriate to each dependency.

## Why

- **Unbounded waits cascade**: One slow dependency without a timeout blocks a thread/goroutine/task, which blocks the caller, which blocks the caller's caller. Within minutes, the entire system is stuck waiting.
- **Naive retries amplify failures**: Retrying immediately and indefinitely turns a momentary blip into a sustained DDoS against your own dependency.
- **Partial failure is normal**: In any system with more than one dependency, at least one will be degraded at any given time. The system must continue to serve users for the features that still work.

## The Patterns

### 1. Timeouts

Set an explicit timeout on every outbound call. No unbounded waits, ever.

| Call type | Typical defaults | Tune to |
|-----------|-----------------|---------|
| HTTP connect | 5s | Network latency to target |
| HTTP read | 30s | Expected response time of the endpoint |
| Database query | 30s | Slowest acceptable query for the use case |
| Connection pool acquire | 5s | How long to wait for a free connection |
| Lock/mutex acquire | 5s | How long contention is acceptable |
| Subprocess execution | 60s | Expected runtime of the child process |

**Deadline propagation**: When a request has a total budget (e.g. 30s), subtract elapsed time before making each downstream call. If 20s have passed, the next call gets at most 10s, not a fresh 30s.

**Log timeout events** with the dependency name, operation, and configured timeout value. This is essential for tuning.

### 2. Retries with Exponential Backoff

Retry transient failures, but do it carefully.

**Retry only when**:
- The error is transient (5xx, connection reset, timeout)
- The operation is idempotent (GET, PUT, DELETE, or operations with idempotency keys)
- The circuit breaker is closed (see below)

**Never retry**:
- Client errors (4xx) — the request is wrong, retrying won't fix it
- Non-idempotent operations without an idempotency key — risk of duplicate side effects
- When the circuit breaker is open — the dependency is confirmed down

**Backoff formula**: `delay = min(base * 2^attempt + jitter, max_delay)`

Sensible defaults:
- Base delay: 100ms
- Max retries: 3
- Max delay: 5s
- Jitter: random ±25% of the calculated delay

Jitter prevents thundering herds — without it, all clients retry at exactly the same moment.

### 3. Circuit Breakers

A circuit breaker tracks failure rate per dependency and stops sending requests when the dependency is confirmed unhealthy.

**Three states**:

```
CLOSED (normal)
  → failures exceed threshold → OPEN
OPEN (rejecting requests)
  → cooldown expires → HALF-OPEN
HALF-OPEN (probing)
  → probe succeeds → CLOSED
  → probe fails → OPEN
```

**Configuration**:
- **Failure threshold**: Open after N consecutive failures, or when error rate exceeds X% over a sliding window (e.g. 50% over 10 requests)
- **Cooldown period**: How long to stay open before probing (e.g. 30s)
- **Probe**: Allow one request through in half-open state. If it succeeds, close. If it fails, re-open.

**Log state transitions** (CLOSED→OPEN, OPEN→HALF-OPEN, HALF-OPEN→CLOSED). Expose circuit state as a metric.

**Interaction with retries**: Retries happen inside the circuit. When the circuit is open, requests fail immediately without retrying — this is the whole point.

### 4. Graceful Degradation

For each external dependency, define what happens when it's unavailable.

| Dependency type | Degradation strategy |
|----------------|---------------------|
| **Hard dependency** (can't function without it) | Return 503, log, alert. Example: primary database. |
| **Soft dependency** (feature degraded without it) | Serve stale cached data, disable the feature, or return partial results. Example: search service, recommendation engine. |

Document dependency criticality explicitly. A soft dependency must never cause a full outage.

**Examples**:
- Search service down → return "search temporarily unavailable" with cached popular results
- Payment provider down → accept the order, queue the payment for later processing
- Analytics service down → drop analytics events silently, continue serving users

### 5. Backpressure and Load Shedding

Prevent resource exhaustion by bounding the amount of work the system accepts.

- **Batch size limits**: Cap items per bulk request (e.g. max 100)
- **Queue depth limits**: Reject new work when the queue is full, rather than growing unboundedly
- **Concurrent connection limits**: Cap outbound connections per dependency
- **Rate limiting on ingress**: Return 429 or 503 when the system is at capacity

Load shedding is deliberate: it's better to reject 10% of requests cleanly than to degrade performance for 100% of requests.

## Composition

These patterns layer on top of each other:

```
Incoming request
  → Backpressure check (reject if overloaded)
    → Circuit breaker check (fail fast if dependency is down)
      → Timeout (bound the wait)
        → Actual call
      → On failure: Retry with backoff (if retriable)
    → On circuit open: Graceful degradation (fallback)
```

## Implementation Notes

### Go

```go
// Timeout: use context deadlines
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

resp, err := client.Do(req.WithContext(ctx))
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        logger.Warn("request timed out", "dependency", "payments", "timeout", "5s")
    }
    return err
}
```

For retries, use a library like `cenkalti/backoff` or write a simple retry loop. For circuit breakers, `sony/gobreaker` is widely used.

Go's `context.Context` supports deadline propagation natively — pass the same context through the call chain and each layer's timeout is bounded by the parent.

### TypeScript

```typescript
// Timeout: use AbortController (native)
const controller = new AbortController();
const timeout = setTimeout(() => controller.abort(), 5000);

try {
  const response = await fetch(url, { signal: controller.signal });
  clearTimeout(timeout);
  return await response.json();
} catch (err) {
  clearTimeout(timeout);
  if (err instanceof DOMException && err.name === "AbortError") {
    logger.warn({ dependency: "payments", timeout: "5s" }, "request timed out");
  }
  throw err;
}
```

For retries, use `p-retry` or `async-retry`. For circuit breakers, use `cockatiel` (supports retry, circuit breaker, bulkhead, timeout — all composable):

```typescript
import { CircuitBreakerPolicy, ConsecutiveBreaker, retry, handleAll, wrap } from "cockatiel";

const circuitBreaker = new CircuitBreakerPolicy(
  handleAll,
  new ConsecutiveBreaker(5),       // open after 5 consecutive failures
  { halfOpenAfter: 30_000 },       // probe after 30s
);

const retryPolicy = retry(handleAll, { maxAttempts: 3, backoff: { type: "exponential", initialDelay: 100 } });

// Compose: retries inside circuit breaker
const policy = wrap(circuitBreaker, retryPolicy);

const result = await policy.execute(() => callPaymentService(payload));
```

### Python

```python
import httpx

# Timeout: httpx has built-in timeout configuration
client = httpx.AsyncClient(
    timeout=httpx.Timeout(connect=5.0, read=30.0, write=10.0, pool=5.0)
)

# Retries: tenacity library
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type

@retry(
    stop=stop_after_attempt(3),
    wait=wait_exponential(multiplier=0.1, max=5),
    retry=retry_if_exception_type((httpx.TimeoutException, httpx.HTTPStatusError)),
)
async def call_payment_service(payload: dict) -> dict:
    response = await client.post("/payments", json=payload)
    response.raise_for_status()
    return response.json()
```

For circuit breakers, `pybreaker` is the standard library. For FastAPI, combine with dependency injection so the circuit breaker is shared across requests.

## When to Bend the Rules

- **Internal calls on the same machine** (localhost): Timeouts are still needed but can be shorter and retries are less critical.
- **Startup-time calls** (loading config from a vault): Retry aggressively — the system can't start without it, and there's no user waiting.
- **Batch/offline processing**: Longer timeouts and more retries are acceptable because latency matters less than completion.
- **Simple scripts**: If a script calls one API and exits, a timeout and a single retry are sufficient. Circuit breakers are for long-running services.
