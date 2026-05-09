# API Design

## Principle

An API is a contract. Design it to be consistent, predictable, and self-documenting. Use standard status codes, structured error responses, pagination for collections, and schema validation on every request. The API should be usable without reading the source code.

## Why

- **Clients are written once and break later**: A frontend, mobile app, or integration written against your API today must still work after your next deploy. Consistency and stability are more important than cleverness.
- **Error handling drives API quality**: A well-designed happy path is easy. What separates good APIs from bad ones is how they communicate failures — wrong input, missing resources, server errors, rate limits.
- **Self-documenting APIs reduce support burden**: When the API returns a clear error with a type URI, a human-readable message, and field-level details, the developer fixes it without filing a support ticket.

## Core Rules

### 1. Use Standard HTTP Status Codes

Don't invent your own codes. Clients already have handling logic for standard codes.

| Code | Meaning | When to use |
|------|---------|-------------|
| **200** | OK | Successful GET, PUT, PATCH |
| **201** | Created | Successful POST that creates a resource |
| **204** | No Content | Successful DELETE or action with no response body |
| **400** | Bad Request | Malformed request (unparseable JSON, wrong content type) |
| **401** | Unauthorized | Missing or invalid authentication credentials |
| **403** | Forbidden | Authenticated but not authorized for this action |
| **404** | Not Found | Resource doesn't exist (or user can't see it — use 404 over 403 to avoid leaking existence) |
| **409** | Conflict | Request conflicts with current state (duplicate, version mismatch) |
| **422** | Unprocessable Entity | Request is well-formed but semantically invalid (business rule violation) |
| **429** | Too Many Requests | Rate limit exceeded — include `Retry-After` header |
| **500** | Internal Server Error | Unexpected server error — never expose internals |

**The 400 vs 422 distinction**: 400 means the request is syntactically broken (can't parse it). 422 means the request is valid JSON/XML but the content violates business rules (e.g. "balance too low").

### 2. Use RFC 9457 Problem Details for Errors

Every error response should follow the [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457) Problem Details format:

```json
{
  "type": "https://api.example.com/errors/insufficient-balance",
  "title": "Insufficient Balance",
  "status": 422,
  "detail": "Account balance is 4.50 EUR, but 10.00 EUR is required.",
  "instance": "/orders/abc-123"
}
```

| Field | Purpose | Stability |
|-------|---------|-----------|
| `type` | Machine-readable error identifier (URI) | Stable — clients match on this |
| `title` | Short human-readable summary | Stable — same for all instances of this error type |
| `status` | HTTP status code (repeated for convenience) | Matches the response status |
| `detail` | Human-readable explanation of this specific occurrence | Varies per occurrence |
| `instance` | URI identifying this specific occurrence | Varies per occurrence |

**Extension fields** for validation errors:

```json
{
  "type": "https://api.example.com/errors/validation",
  "title": "Validation Error",
  "status": 422,
  "detail": "Request body has 2 validation errors",
  "errors": [
    {"field": "email", "message": "must be a valid email address"},
    {"field": "age", "message": "must be a positive integer"}
  ]
}
```

**Content-Type**: `application/problem+json`

### 3. Validate Every Request

Validate request body, query parameters, and path parameters at the handler layer. Reject invalid input immediately with 400 or 422 and field-level error details.

**Validate**:
- Presence of required fields
- Types (string, number, boolean)
- Constraints (min/max length, ranges, patterns)
- Business rules (valid enum values, existing references)

**Return field-level errors** so clients know exactly what to fix — not just "invalid request."

### 4. Paginate Collections

Every list endpoint must support pagination. Never return unbounded collections.

**Cursor-based** (preferred for large or frequently-updated datasets):
```json
{
  "items": [...],
  "next_cursor": "eyJpZCI6MTAwfQ==",
  "has_more": true
}
```

**Offset-based** (acceptable for small, stable datasets):
```json
{
  "items": [...],
  "total": 243,
  "page": 2,
  "page_size": 20
}
```

Cursor-based pagination is more performant (no `OFFSET` query) and handles insertions/deletions between pages correctly. Offset-based is simpler but can skip or duplicate items if data changes between requests.

### 5. Rate Limit with Standard Headers

When rate limiting is active, include these headers on *every* response (not just 429s):

```
RateLimit-Limit: 100
RateLimit-Remaining: 42
RateLimit-Reset: 1625000000
```

When the limit is exceeded, return 429 with a `Retry-After` header:

```
HTTP/1.1 429 Too Many Requests
Retry-After: 30
Content-Type: application/problem+json

{
  "type": "https://api.example.com/errors/rate-limited",
  "title": "Rate Limit Exceeded",
  "status": 429,
  "detail": "100 requests per minute allowed. Try again in 30 seconds."
}
```

This allows well-behaved clients to self-throttle before hitting the limit.

### 6. Generate Machine-Readable Documentation

Generate an OpenAPI 3.x specification from code annotations or maintain a spec-first definition. Serve interactive documentation (Swagger UI, Redoc) in non-production environments.

**Validate in CI** that the spec matches the implementation — use contract tests or spec linting (Spectral, optic) to catch drift.

### 7. Never Expose Internals in Errors

Production error responses must never contain:
- Stack traces
- SQL queries or database error messages
- Internal file paths
- Dependency names or versions
- Raw exception messages from third-party libraries

Log these server-side with full detail. Return a generic message to the client.

```
BAD:  {"error": "psycopg2.errors.UniqueViolation: duplicate key value violates unique constraint \"users_email_key\""}
GOOD: {"type": "/errors/conflict", "title": "Conflict", "status": 409, "detail": "A user with this email already exists."}
```

## Implementation Notes

### Go

```go
// Problem Details response helper
type ProblemDetails struct {
    Type     string `json:"type"`
    Title    string `json:"title"`
    Status   int    `json:"status"`
    Detail   string `json:"detail,omitempty"`
    Instance string `json:"instance,omitempty"`
}

func writeProblem(w http.ResponseWriter, status int, title, detail string) {
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ProblemDetails{
        Type:   fmt.Sprintf("/errors/%s", strings.ReplaceAll(strings.ToLower(title), " ", "-")),
        Title:  title,
        Status: status,
        Detail: detail,
    })
}
```

### TypeScript (Express)

```typescript
import { Request, Response, NextFunction } from "express";
import { ZodError } from "zod";

// Zod validation error → RFC 9457 Problem Details
app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  if (err instanceof ZodError) {
    res.status(422).json({
      type: "/errors/validation",
      title: "Validation Error",
      status: 422,
      detail: `${err.issues.length} validation error(s)`,
      errors: err.issues.map((issue) => ({
        field: issue.path.join("."),
        message: issue.message,
      })),
    });
  } else if (err instanceof DomainError) {
    const status = domainErrorToStatus(err);
    res.status(status).json({
      type: `/errors/${err.code.toLowerCase().replace(/_/g, "-")}`,
      title: err.message,
      status,
      detail: err.message,
    });
  } else {
    logger.error({ err, path: req.path }, "unexpected error");
    res.status(500).json({
      type: "/errors/internal",
      title: "Internal error",
      status: 500,
      detail: "An unexpected error occurred",
    });
  }
});
```

For NestJS, use exception filters. For Fastify, use `setErrorHandler`. Zod integrates with both via middleware or pipes.

### Python (FastAPI)

```python
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError

@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    return JSONResponse(
        status_code=422,
        content={
            "type": "/errors/validation",
            "title": "Validation Error",
            "status": 422,
            "detail": f"{len(exc.errors())} validation error(s)",
            "errors": [
                {"field": ".".join(str(l) for l in e["loc"]), "message": e["msg"]}
                for e in exc.errors()
            ],
        },
        headers={"Content-Type": "application/problem+json"},
    )
```

FastAPI with Pydantic automatically validates request bodies against the model schema. Override the default validation error handler to return RFC 9457 format instead of FastAPI's default.

## When to Bend the Rules

- **Internal APIs between services you control**: Simpler error formats are fine if both sides agree. Still use standard status codes.
- **GraphQL APIs**: Status codes and pagination work differently. The error principles (structured, machine-readable, no internal leaks) still apply.
- **Streaming/WebSocket APIs**: REST conventions don't apply to persistent connections. Define your own message format for errors, but keep the same principles (type, detail, no internals).
- **MVP/prototype**: Skip OpenAPI generation. Don't skip error handling — bad error responses in a prototype become bad error responses in production.
- **Security-sensitive endpoints**: Sometimes detailed error messages create an attack surface. Login endpoints should return the same error for "user not found" and "wrong password" — otherwise attackers can enumerate valid accounts. Apply the same logic to password reset, invitation codes, and any endpoint where confirming or denying the existence of a resource helps an attacker. Combine with aggressive rate limiting on these endpoints — even obfuscated responses leak information at scale when an attacker can make thousands of attempts per minute.
