# Structured Logging

## Principle

Log structured data (key-value pairs), not formatted strings. Use module-scoped loggers, appropriate severity levels, and consistent field names. Never log secrets. Always include context that helps answer "what happened, where, and to whom?"

## Why

- **Queryable**: Structured fields enable filtering ("show all errors for user X in the last hour") without regex parsing.
- **Consistent**: Standardised field names across the codebase mean dashboards and alerts work reliably.
- **Efficient**: Lazy evaluation avoids formatting strings for messages that are never emitted at the current log level.
- **Debuggable**: When an incident happens at 3 AM, good logs are the difference between a 10-minute fix and a 2-hour investigation.

## Core Rules

### 1. Use Module-Scoped Loggers

Create one logger per module/package, named after that module. This enables per-module level configuration and makes it immediately clear where a log line originated.

```
GOOD: Logger named "app.service.user" → clearly from the user service
BAD:  Global root logger → origin unknown, can't filter by module
```

### 2. Use Structured Fields, Not String Interpolation

Attach context as key-value pairs, not embedded in the message string.

```
GOOD: Log message "user login successful" with fields {user_id: "abc", method: "oauth"}
BAD:  Log message "User abc logged in via oauth"
```

The first version lets you query `user_id = "abc"` across all log types. The second requires regex parsing and breaks when the format changes.

### 3. Choose the Right Severity Level

| Level | When | Examples |
|-------|------|----------|
| **Debug** | Diagnostic detail for developers. Disabled in production. | Function entry/exit, variable values, query parameters |
| **Info** | Normal operations worth recording. | Server started, request completed, job finished |
| **Warn** | Unexpected but handled. System continues normally. | Retry succeeded, deprecated API called, fallback used, client error (4xx) |
| **Error** | Operation failed. Requires attention. | Database query failed, external API unreachable, server error (5xx) |
| **Fatal/Critical** | System cannot continue. | Out of memory, config file missing, port already in use |

**Common mistakes:**
- Logging successful operations at Error level ("User logged in" is Info, not Error)
- Logging actual failures at Warn level ("Database connection failed" is Error, not Warn)
- Using Debug in production (overwhelms log storage and obscures real issues)

### 4. Include Exception/Error Context

When logging an error, always include the stack trace or error chain. The message alone is rarely sufficient for debugging.

```
GOOD: Log the error with full stack trace / error chain attached
BAD:  Log only the error message string, discarding the trace
```

### 5. Use Consistent Field Names

Define a standard set of field names and use them everywhere. Document them.

| Field | Type | Description |
|-------|------|-------------|
| `user_id` | string | User identifier (always stringified, never the raw object) |
| `request_id` | string | Request correlation ID for tracing |
| `session_id` | string | Session identifier |
| `error_type` | string | Error class/type name |
| `duration_ms` | number | Operation duration in milliseconds |
| `operation` | string | What was being attempted |

Inconsistent naming (`userId` vs `user_id` vs `userID`) breaks queries and dashboards.

### 6. Never Log Secrets

**Never log**: passwords, API keys, tokens, session secrets, PII (SSN, full credit card numbers), private keys.

**Do log**: boolean indicators of whether a secret is present.

```
GOOD: {has_api_key: true, has_endpoint: true}
BAD:  {api_key: "sk-abc123..."}
```

### 7. Truncate Large Values

Log previews of large values, not the full content. This controls storage costs and keeps logs readable.

```
GOOD: {query_preview: "SELECT * FROM users WHERE...", query_length: 2048}
BAD:  {query: "<2048 characters of SQL>"}
```

### 8. Use Lazy Evaluation

Don't compute log message content if the log level means it won't be emitted. Most logging libraries support this natively.

```
GOOD: Logger evaluates the message only if Debug level is enabled
BAD:  Expensive string formatting runs on every call, even when Debug is disabled
```

### 9. Log at System Boundaries

The most valuable log lines are at the edges: incoming requests, outgoing calls, and operation results. Log:

- **Request received**: method, path, client ID (not the full body)
- **Request completed**: status, duration, response size
- **External call made**: target service, operation, duration, success/failure
- **Background job started/completed**: job type, item count, duration, success/failure count

Internal function calls generally don't need logging unless they're complex business operations.

## Async/Concurrent Applications

In async or multi-threaded applications, logging can block the event loop or create contention. Two patterns to address this:

### Non-Blocking Logging

Route log records through a queue. A background thread/goroutine handles the actual I/O (writing to files, sending to log aggregation services).

```
[Application code]
    ↓ logger.info() — fast, non-blocking
[In-memory queue]
    ↓ background worker
[File / Network I/O] — blocking, but off the hot path
```

### Correlation IDs

In concurrent systems, attach a request ID or trace ID to every log line so you can reconstruct the sequence of events for a single request across goroutines/tasks/threads.

Store the correlation ID in request-scoped context (Go's `context.Context`, Rust's `tracing::Span`, Python's `contextvars`) and inject it into every log line automatically via middleware or logging configuration.

## Implementation Notes

### Go

Use `log/slog` (standard library, Go 1.21+). It provides structured logging with levels, groups, and pluggable handlers.

```go
package user

import "log/slog"

// Module-scoped logger
var logger = slog.Default().With("module", "user")

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        logger.ErrorContext(ctx, "failed to get user",
            "user_id", id.String(),
            "error", err,
        )
        return nil, fmt.Errorf("get user %s: %w", id, err)
    }

    logger.InfoContext(ctx, "user retrieved",
        "user_id", id.String(),
    )
    return user, nil
}
```

For request correlation, use middleware that injects a `slog.Logger` with `request_id` into the context.

### TypeScript

Use [pino](https://getpino.io/) — it's the fastest structured logger for Node.js and outputs JSON by default.

```typescript
import pino from "pino";

// Module-scoped logger
const logger = pino({ name: "user-service" });

async function getById(id: string): Promise<User> {
  try {
    const user = await repo.getById(id);
    logger.info({ userId: id }, "user retrieved");
    return user;
  } catch (err) {
    logger.error({ userId: id, err }, "failed to get user");
    throw err;
  }
}
```

For request correlation, use `pino-http` middleware which automatically logs request/response and attaches a `reqId` to every log line. In Express:

```typescript
import pinoHttp from "pino-http";

app.use(pinoHttp({
  genReqId: (req) => req.headers["x-request-id"] || crypto.randomUUID(),
}));
```

Use `pino-pretty` for human-readable output in development, raw JSON in production.

### Python

Use the standard `logging` module with `getLogger(__name__)`. For structured fields, use the `extra` parameter. For production JSON output, add `structlog` or `python-json-logger`.

```python
import logging

logger = logging.getLogger(__name__)

async def get_by_id(self, user_id: UUID) -> User:
    try:
        user = await self.repo.get_by_id(user_id)
    except Exception as e:
        logger.error(
            "Failed to get user",
            extra={"user_id": str(user_id), "error_type": type(e).__name__},
            exc_info=True,
        )
        raise

    logger.info("User retrieved", extra={"user_id": str(user_id)})
    return user
```

For async applications (FastAPI, aiohttp), use `QueueHandler` + `QueueListener` to avoid blocking the event loop on log I/O.

For request correlation, use `contextvars.ContextVar` with a logging `Filter` that injects the request ID into every log record.

## Environment-Based Configuration

Configure logging behaviour through environment variables, not code changes:

| Variable | Purpose | Example |
|----------|---------|---------|
| `LOG_LEVEL` | Root log level | `INFO` (production), `DEBUG` (development) |
| `LOG_FORMAT` | Output format | `json` (production), `text` (development) |
| `LOG_MODULE_LEVELS` | Per-module overrides | `app.repo=WARN,app.service=DEBUG` |

This allows operators to increase verbosity for debugging without redeploying.

## When to Bend the Rules

- **Local development**: Human-readable text format is fine. Switch to JSON for production.
- **High-throughput data pipelines**: Sample logs (log 1% of events) rather than logging everything. Metrics are better than logs for high-cardinality counters.
- **Short-lived scripts**: `fmt.Println` or `print()` is fine for a script that runs once and exits. Add structured logging when the script becomes a service.
