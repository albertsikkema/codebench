# Error Handling

## Principle

Errors are not exceptional — they are a normal part of program execution. Handle them explicitly at every layer, propagate them with context, translate them at boundaries, and never swallow them silently. The goal is that every failure produces enough information to diagnose the problem without reproducing it.

## Why

- **Silent failures are the worst failures**: A swallowed exception means the system continues in an inconsistent state. The user sees wrong results, not an error. The bug is discovered hours or days later, with no trail to follow.
- **Generic errors are useless**: `"something went wrong"` in the logs means a 2 AM debugging session. `"payment gateway returned 503 for order abc-123 after 3 retries (timeout: 5s)"` means a 2-minute diagnosis.
- **Users and machines need different error formats**: Humans need a clear message. API clients need a structured, stable error code they can branch on. Internal logs need the stack trace. One error, three representations.

## Core Rules

### 1. Never Swallow Errors

Every error must be either **handled** (the code takes corrective action) or **propagated** (the caller deals with it). Logging and continuing is not handling — it's ignoring with a paper trail.

```
BAD:  catch (e) { log(e) }              // then what? State is corrupt.
BAD:  catch (e) { return defaultValue }  // caller doesn't know it failed
BAD:  if err != nil { return nil }       // error disappears

GOOD: catch (e) { log(e); throw }       // propagate after logging
GOOD: catch (e) { return fallback, ErrDegraded }  // caller knows it's degraded
GOOD: if err != nil { return fmt.Errorf("load config: %w", err) }  // wrap and propagate
```

### 2. Add Context When Propagating

Each layer that propagates an error should add what it was doing. The final error message reads like a stack of explanations.

```
Final error: "create order: charge payment: POST /payments: connection refused"

Each layer added context:
  - handler:    "create order"
  - service:    "charge payment"
  - http client: "POST /payments"
  - network:    "connection refused"
```

### 3. Translate Errors at Boundaries

Each architectural boundary (handler ↔ service, service ↔ repository, system ↔ user) should translate errors into the vocabulary of the outer layer.

| Boundary | Inner error | Outer error |
|----------|-------------|-------------|
| Repository → Service | `sql.ErrNoRows` | `ErrUserNotFound` |
| Service → Handler | `ErrUserNotFound` | HTTP 404 |
| Service → Handler | `ErrInsufficientBalance` | HTTP 422 with problem details |
| Service → Handler | Unexpected panic/exception | HTTP 500 with generic message |

**Never expose internal details to external callers**: No stack traces, no SQL errors, no file paths, no dependency names in user-facing responses.

### 4. Use Structured Error Responses

API errors should follow a consistent, machine-parseable format. RFC 9457 (Problem Details for HTTP APIs) is the standard:

```json
{
  "type": "https://example.com/errors/insufficient-balance",
  "title": "Insufficient Balance",
  "status": 422,
  "detail": "Account balance is 4.50, but 10.00 is required.",
  "instance": "/orders/abc-123"
}
```

The `type` field is a stable identifier that clients can match on. The `detail` field is human-readable. Extension fields (e.g. `errors: [...]` for field-level validation) are added as needed.

### 5. Log Errors with Full Context

When an error is logged (typically at the boundary where it's translated for the caller), include:

- **What was being attempted** (operation name)
- **Who triggered it** (user ID, request ID)
- **What went wrong** (error type, message)
- **The full stack trace or error chain**
- **Relevant identifiers** (order ID, resource ID)

See the structured logging best practice for field naming conventions.

### 6. Distinguish Retriable from Terminal Errors

Not all errors deserve the same treatment:

| Error type | Example | Action |
|-----------|---------|--------|
| **Retriable** | Timeout, 503, connection reset | Retry with backoff |
| **Client error** | 400, 422, validation failure | Return to caller, do not retry |
| **Terminal** | 401, 403, configuration missing | Fail immediately, alert if unexpected |
| **Corruption** | Data integrity violation | Fail, alert, investigate |

Make this distinction explicit in your error types so callers can branch on it.

### 7. Handle Errors in Loops Deliberately

When processing a collection, decide upfront which failure strategy applies:

- **Abort**: When items are dependent (ordered pipeline, sequential steps where later items depend on earlier ones)
- **Continue**: When items are independent (batch import, sending notifications)
- **Revert all**: When partial completion is worse than total failure (financial transactions, multi-record updates that must be atomic). Wrap the entire batch in a transaction so all changes roll back on any failure

When continuing, log each failure with the item identifier and track success/failure counts. See the batch error handling best practice for the full pattern.

## Anti-Patterns

### Catch-All at the Wrong Level

```
BAD:  Wrapping the entire request handler in try/catch and returning 500 for everything.
      This hides whether the error is a validation failure (400), auth failure (401),
      or actual server error (500).

GOOD: Let specific errors propagate to a global error handler that maps error types
      to appropriate status codes.
```

### Error Logging at Every Layer

```
BAD:  Repository logs the error. Service logs the error. Handler logs the error.
      The same error appears 3 times in the logs.

GOOD: Propagate with context at each layer. Log once at the outermost boundary.
      Inner layers may log at DEBUG level for development, but not at ERROR.
```

### Returning Nil/Null/None on Error

```
BAD:  func GetUser(id) -> User? { try { ... } catch { return nil } }
      Caller can't distinguish "user not found" from "database is down."

GOOD: func GetUser(id) -> (User, error)     // Go
GOOD: fn get_user(id) -> Result<User, Error> // Rust
GOOD: raise UserNotFound / raise DatabaseError  // Python
```

## Implementation Notes

### Go

Go's explicit error handling makes these patterns natural:

```go
// Wrap with context at each layer
func (s *OrderService) Create(ctx context.Context, req CreateOrderReq) (*Order, error) {
    user, err := s.users.GetByID(ctx, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("create order: get user: %w", err)
    }

    // ...
}

// Translate at the handler boundary
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    order, err := h.service.Create(r.Context(), req)
    if err != nil {
        switch {
        case errors.Is(err, domain.ErrUserNotFound):
            writeProblem(w, 404, "User not found", err.Error())
        case errors.Is(err, domain.ErrInsufficientBalance):
            writeProblem(w, 422, "Insufficient balance", err.Error())
        default:
            slog.ErrorContext(r.Context(), "unexpected error", "error", err)
            writeProblem(w, 500, "Internal error", "An unexpected error occurred")
        }
        return
    }
}
```

Use sentinel errors (`var ErrNotFound = errors.New(...)`) or custom error types for domain errors. Use `%w` for wrapping. Use `errors.Is()` and `errors.As()` for matching.

### TypeScript

Define domain errors as typed classes. Use a global error handler to translate at the boundary.

```typescript
// Domain errors
class DomainError extends Error {
  constructor(message: string, public readonly code: string) {
    super(message);
    this.name = "DomainError";
  }
}

class UserNotFoundError extends DomainError {
  constructor(public readonly userId: string) {
    super(`User ${userId} not found`, "USER_NOT_FOUND");
  }
}

class InsufficientBalanceError extends DomainError {
  constructor(public readonly have: number, public readonly need: number) {
    super(`Insufficient balance: have ${have}, need ${need}`, "INSUFFICIENT_BALANCE");
  }
}

// Service: wrap with context
async function createOrder(req: CreateOrderReq): Promise<Order> {
  const user = await userRepo.getById(req.userId);
  if (!user) throw new UserNotFoundError(req.userId);
  // ...
}

// Handler: translate at the boundary (Express)
app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  if (err instanceof UserNotFoundError) {
    res.status(404).json({
      type: "/errors/user-not-found",
      title: "User not found",
      status: 404,
      detail: err.message,
    });
  } else if (err instanceof InsufficientBalanceError) {
    res.status(422).json({
      type: "/errors/insufficient-balance",
      title: "Insufficient Balance",
      status: 422,
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

For NestJS, use exception filters. For Fastify, use `setErrorHandler`. The pattern is the same: domain errors map to specific status codes, everything else is 500 with a generic message.

### Python

```python
# Define domain exceptions
class DomainError(Exception):
    """Base for all domain errors."""

class UserNotFound(DomainError):
    def __init__(self, user_id: UUID):
        self.user_id = user_id
        super().__init__(f"User {user_id} not found")

class InsufficientBalance(DomainError):
    def __init__(self, have: float, need: float):
        self.have = have
        self.need = need
        super().__init__(f"Insufficient balance: have {have}, need {need}")

# Translate at the handler boundary (FastAPI)
@app.exception_handler(UserNotFound)
async def user_not_found_handler(request: Request, exc: UserNotFound):
    return JSONResponse(status_code=404, content={
        "type": "/errors/user-not-found",
        "title": "User not found",
        "status": 404,
        "detail": str(exc),
    })

@app.exception_handler(Exception)
async def unexpected_error_handler(request: Request, exc: Exception):
    logger.error("Unexpected error", exc_info=True, extra={"path": request.url.path})
    return JSONResponse(status_code=500, content={
        "type": "/errors/internal",
        "title": "Internal error",
        "status": 500,
        "detail": "An unexpected error occurred",
    })
```

Use exception hierarchies for domain errors. Use FastAPI's `exception_handler` decorator or middleware for boundary translation. Always log with `exc_info=True` to capture the traceback.

## When to Bend the Rules

- **Fire-and-forget operations** (analytics, non-critical logging): Swallowing errors is acceptable when the operation is truly optional and failure has no user impact. Still log at DEBUG or WARN level.
- **Best-effort cleanup** (closing connections in a finally block): If cleanup fails, log it but don't override the original error.
- **Panic/unrecoverable errors**: Some errors (out of memory, stack overflow) can't be handled gracefully. Let the process crash and rely on the orchestrator to restart it.
