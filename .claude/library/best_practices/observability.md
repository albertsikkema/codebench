# Observability

## Principle

A system is observable when you can understand its internal state from its external outputs — logs, metrics, and traces. These three signals are complementary: logs tell you what happened, metrics tell you how much, and traces tell you the journey of a single request through the system.

## Why

- **You can't debug what you can't see**: In production, you can't attach a debugger. Observability is your debugger.
- **Alerts without context are noise**: "Error rate is 5%" is an alert. "Error rate is 5%, all from the `/payments` endpoint, all returning `connection refused` to the Stripe API, started 3 minutes ago" is a diagnosis. The difference is observability.
- **Mean Time to Recovery (MTTR) is what matters**: You can't prevent all outages. You can make them short. Good observability turns a 2-hour investigation into a 5-minute one.

## The Three Pillars

### 1. Structured Logging

Covered in detail in the **Structured Logging** best practice. Key points:

- Log structured key-value pairs, not formatted strings
- Include request ID, user ID, and operation context in every log line
- Use appropriate severity levels
- Never log secrets

Logs answer: **"What happened?"**

### 2. Metrics

Metrics are numeric measurements aggregated over time. They tell you the health of the system at a glance.

**The RED Method** (for request-driven services):

| Metric | What it measures | Example |
|--------|-----------------|---------|
| **Rate** | Requests per second | `http_requests_total` |
| **Errors** | Failed requests per second | `http_errors_total` |
| **Duration** | Time per request (histogram) | `http_request_duration_seconds` |

Export these per endpoint with labels for method, path, and status code.

**The USE Method** (for resources):

| Metric | What it measures | Example |
|--------|-----------------|---------|
| **Utilization** | How busy is it? (0-100%) | CPU usage, DB pool usage |
| **Saturation** | How full is the queue? | Queue depth, connection waiters |
| **Errors** | How often does it fail? | Connection failures, OOM events |

**Application-specific metrics**:
- Database connection pool: size, in-use, idle, wait time
- Message queue: depth, processing rate, dead letters
- Cache: hit rate, miss rate, eviction rate
- Business metrics: signups/hour, orders/minute, revenue/day

### 3. Distributed Tracing

A trace follows a single request through every service and component it touches. Each step is a **span** with a start time, duration, and metadata.

```
[Client] → [API Gateway] → [Order Service] → [Payment Service]
                                           → [Inventory Service]
                                           → [Notification Service]

Trace ID: abc-123
├── Span: API Gateway (2ms)
├── Span: Order Service (45ms)
│   ├── Span: Validate order (3ms)
│   ├── Span: Payment Service call (30ms)
│   └── Span: Inventory Service call (8ms)
└── Span: Notification Service (fire-and-forget)
```

**Request ID / Correlation ID**:
- Generate a UUID for each incoming request
- Pass it in the `X-Request-ID` header to all downstream calls
- Include it in every log line for that request
- This is the minimum viable tracing — even without a tracing system, correlation IDs let you reconstruct request flows from logs

## Core Rules

### 1. Every Request Gets a Correlation ID

Generate a UUID at the entry point (API gateway, load balancer, or first handler). Propagate it through:
- HTTP headers (`X-Request-ID`)
- Log fields (`request_id`)
- Database queries (as a comment or in audit context)
- Message queue headers

If the incoming request already has an `X-Request-ID`, use it (the upstream caller is part of the same trace).

### 2. Log at System Boundaries

The most valuable log lines are at the edges:

| Boundary | What to log |
|----------|------------|
| **Request received** | Method, path, client identifier |
| **Request completed** | Status code, duration, response size |
| **Outbound call made** | Target service, operation, duration, success/failure |
| **Background job started** | Job type, input parameters |
| **Background job completed** | Duration, items processed, success/failure count |

Internal function calls generally don't need logging unless they represent significant business operations.

### 3. Track the Four Golden Signals

For every service, track:
1. **Latency**: Response time distribution (p50, p95, p99)
2. **Traffic**: Request rate
3. **Errors**: Error rate and error types
4. **Saturation**: Resource utilization and queue depths

These four signals are sufficient to detect most problems.

### 4. Define Alerts with Severity Levels

| Severity | Response time | Example thresholds |
|----------|--------------|-------------------|
| **Critical** | Page immediately | Error rate > 5% for 5 min, health check failing |
| **Warning** | Investigate within hours | p95 latency > 2x target for 10 min, disk 80% full |
| **Info** | Review next business day | Elevated error rate in non-critical path, cert expiring in 30 days |

**Alert quality matters**: Too many alerts = alert fatigue = missed real alerts. Every alert should be actionable — if you can't do anything about it, it shouldn't page you.

### 5. Store Dashboards as Code

Dashboard definitions should be version-controlled (Grafana JSON, Terraform, Pulumi, or equivalent). This ensures:
- Dashboards are reproducible after infrastructure changes
- Dashboard changes are reviewed like code changes
- New environments get the same dashboards automatically

**Standard dashboards**:
1. **Service overview**: Request rate, error rate, latency (RED)
2. **Resources**: Connection pools, queue depths, memory, CPU (USE)
3. **Business metrics**: Key user flows, conversion rates, revenue

### 6. Separate Health from Readiness

| Endpoint | Purpose | What it checks | Who calls it |
|----------|---------|---------------|-------------|
| `GET /health` | Is the process alive? | Process is listening | Load balancer, uptime monitor |
| `GET /ready` | Can it serve traffic? | DB connected, dependencies available | Orchestrator (K8s readiness probe) |

The health endpoint must be fast, cheap, and never fail unless the process is truly dead. The readiness endpoint checks dependencies and can return 503 during startup or when a dependency is down.

**Security**: The health endpoint can be public. The readiness endpoint should be restricted (it reveals dependency information).

## Implementation Notes

### Go

```go
// Middleware: request ID and request logging
func RequestLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.NewString()
        }

        ctx := context.WithValue(r.Context(), "request_id", requestID)
        logger := slog.With("request_id", requestID, "method", r.Method, "path", r.URL.Path)

        start := time.Now()
        ww := &responseWriter{ResponseWriter: w, status: 200}

        next.ServeHTTP(ww, r.WithContext(ctx))

        logger.Info("request completed",
            "status", ww.status,
            "duration_ms", time.Since(start).Milliseconds(),
        )
        w.Header().Set("X-Request-ID", requestID)
    })
}
```

For metrics, use `prometheus/client_golang`. For tracing, use `go.opentelemetry.io/otel`.

### TypeScript (Express)

```typescript
import { randomUUID } from "crypto";
import pino from "pino";
import pinoHttp from "pino-http";

const logger = pino();

// Middleware: request ID, logging, and timing
const httpLogger = pinoHttp({
  genReqId: (req) => (req.headers["x-request-id"] as string) || randomUUID(),
  customSuccessMessage: (req, res) => `${req.method} ${req.url} completed`,
  customErrorMessage: (req, res, err) => `${req.method} ${req.url} failed`,
  customProps: (req) => ({
    requestId: req.id,
  }),
});

app.use(httpLogger);

// Health and readiness endpoints
app.get("/health", (req, res) => res.json({ status: "ok" }));

app.get("/ready", async (req, res) => {
  try {
    await prisma.$queryRaw`SELECT 1`;
    res.json({ status: "ok", components: { database: "ok" } });
  } catch {
    res.status(503).json({ status: "error", components: { database: "error" } });
  }
});
```

For metrics, use `prom-client` with `express-prom-bundle` or `fastify-metrics`. For tracing, use `@opentelemetry/sdk-node` with auto-instrumentation.

### Python (FastAPI)

```python
import time
import uuid
from fastapi import Request

@app.middleware("http")
async def observability_middleware(request: Request, call_next):
    request_id = request.headers.get("x-request-id", str(uuid.uuid4()))
    start = time.monotonic()

    # Store in context for downstream logging
    request.state.request_id = request_id

    response = await call_next(request)

    duration_ms = (time.monotonic() - start) * 1000
    logger.info(
        "Request completed",
        extra={
            "request_id": request_id,
            "method": request.method,
            "path": request.url.path,
            "status": response.status_code,
            "duration_ms": round(duration_ms, 2),
        },
    )
    response.headers["X-Request-ID"] = request_id
    return response
```

For metrics, use `prometheus-fastapi-instrumentator` or the `prometheus_client` library directly. For tracing, use `opentelemetry-api` + `opentelemetry-sdk`.

## When to Bend the Rules

- **Simple scripts and CLIs**: Structured logging to stderr is sufficient. No metrics or tracing needed.
- **Internal tools with low traffic**: Logging + basic health check is enough. Add metrics when you need to know "how much" and "how fast."
- **Early-stage prototypes**: Start with structured logging and a health endpoint. Add metrics and tracing when the system complexity justifies the tooling cost.
