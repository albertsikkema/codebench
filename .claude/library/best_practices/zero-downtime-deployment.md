# Zero-Downtime Deployment

## Principle

Users should never see an error caused by a deployment. Old and new versions of the application must coexist during rollout without data corruption or request failures. This requires backward-compatible code, expand-contract migrations, and health-check-gated rollouts.

## Why

- **Deployments are the most common cause of outages**: More incidents start with "we just deployed" than with any other trigger. Zero-downtime deployment doesn't eliminate deployment risk, but it eliminates the class of errors caused by the transition itself.
- **"Just do it during off-hours" doesn't scale**: There are no off-hours for global services. Even for regional services, maintenance windows create pressure to rush and skip validation.
- **Rollback must be instant**: If the new version has a bug, reverting to the old version must work without a second migration or data fix.

## The Core Constraint

During a rolling deployment, both old and new versions of the application run simultaneously, sharing the same database. Every change must be compatible with both versions.

```
Time →
─────────────────────────────────────
Old v1  ████████████░░░░░░░░░░░░░░░░░
New v2  ░░░░░░░░░░░████████████████████
                   ↑ overlap window ↑
                   Both versions serve traffic
                   Both versions read/write the same database
```

## Core Rules

### 1. Expand-Contract Migrations

Never make breaking schema changes in a single step. Use four phases:

**Phase 1 — Expand**: Add new columns or tables. Set defaults. Don't drop or rename anything.

```sql
-- EXPAND: Add new column with default
ALTER TABLE users ADD COLUMN display_name VARCHAR(255) DEFAULT '';
```

**Phase 2 — Migrate data**: Backfill new structures from old data.

```sql
-- BACKFILL: Populate new column from existing data
UPDATE users SET display_name = first_name || ' ' || last_name WHERE display_name = '';
```

**Phase 3 — Deploy new code**: Application reads from and writes to new structures. Old columns are still present (old code can still use them during rollout overlap).

**Phase 4 — Contract**: Drop old columns in a separate migration, deployed *after* all instances run the new code.

```sql
-- CONTRACT: Only after all instances use display_name
ALTER TABLE users DROP COLUMN first_name;
ALTER TABLE users DROP COLUMN last_name;
```

**Rules**:
- Never add `NOT NULL` without a default value
- Never drop or rename in the same migration that adds the replacement
- Phases 1–2 and Phase 4 are separate deployments, not one

### 2. Backward-Compatible Application Code

During the overlap window, old code runs against a database that may have new columns, and new code runs against a database that may still have old columns.

**Safe changes**:
- Adding a new optional field/column
- Adding a new API endpoint
- Adding a new enum value (if old code ignores unknown values)
- Changing internal implementation without changing the interface

**Unsafe changes** (require expand-contract):
- Renaming a column or field
- Removing a column, field, or endpoint
- Changing a field's type
- Adding a required field without a default
- Changing the meaning of an existing enum value

### 3. Health-Check-Gated Rollouts

The orchestrator (Kubernetes, Docker Swarm, load balancer) should only route traffic to instances that pass health checks.

**Two endpoints**:

| Endpoint | Purpose | Checks | Access |
|----------|---------|--------|--------|
| `GET /health` | Is the process alive? | Process is running, listening | Public |
| `GET /ready` | Can it serve traffic? | Database connected, dependencies available | Internal |

**Rollout sequence**:
1. Start new instance
2. Wait for readiness check to pass
3. Add to load balancer
4. Drain old instance (stop sending new requests)
5. Wait for in-flight requests to complete (with timeout)
6. Terminate old instance

Old instances are never killed before new ones are ready. If the new version fails readiness checks, the rollout stops.

### 4. Graceful Shutdown

When an instance receives SIGTERM (the shutdown signal), it must:

1. Stop accepting new connections
2. Finish in-flight requests (with a timeout, e.g. 30s)
3. Close database connections and release resources
4. Exit with code 0

```
SIGTERM received
  → Stop accepting new requests
  → Wait for in-flight requests (max 30s)
  → Close DB connections
  → Close message queue consumers
  → Flush logs
  → Exit 0
```

If in-flight requests don't complete within the timeout, force-exit. The timeout must be shorter than the orchestrator's kill timeout (Kubernetes `terminationGracePeriodSeconds`).

### 5. Migration Rollback Scripts

Every migration must have both an up and a down script. The down script must be:
- **Tested**: Run in CI, not just written
- **Idempotent**: Safe to run multiple times without error
- **Data-preserving**: Don't drop data that wasn't created by the up script

```sql
-- UP
ALTER TABLE orders ADD COLUMN tracking_number VARCHAR(100);

-- DOWN
ALTER TABLE orders DROP COLUMN IF EXISTS tracking_number;
```

## Deployment Strategies

| Strategy | How it works | Rollback speed | Complexity |
|----------|-------------|----------------|------------|
| **Rolling** | Replace instances one at a time | Moderate (deploy old version) | Low |
| **Blue-Green** | Run two full environments, switch traffic | Instant (switch back) | Medium |
| **Canary** | Route small % of traffic to new version | Instant (route all to old) | High |

Rolling deployment is the default for most applications. Blue-green is better when you need instant rollback. Canary is for high-traffic services where you want to validate with real traffic before full rollout.

## Implementation Notes

### Go

```go
// Graceful shutdown
func main() {
    srv := &http.Server{Addr: ":8080", Handler: router}

    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    // Wait for SIGTERM
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
    <-quit

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("forced shutdown: %v", err)
    }

    // Close other resources
    db.Close()
    log.Println("server stopped")
}
```

### TypeScript (Node.js)

```typescript
import { createServer } from "http";

const server = createServer(app);
server.listen(8080);

// Graceful shutdown
function shutdown(signal: string) {
  console.log(`${signal} received, shutting down gracefully`);

  server.close(async () => {
    // Close database connections
    await prisma.$disconnect();
    // Close other resources
    await redis.quit();
    console.log("shutdown complete");
    process.exit(0);
  });

  // Force exit if graceful shutdown takes too long
  setTimeout(() => {
    console.error("forced shutdown after timeout");
    process.exit(1);
  }, 30_000);
}

process.on("SIGTERM", () => shutdown("SIGTERM"));
process.on("SIGINT", () => shutdown("SIGINT"));
```

For NestJS, use the built-in `app.enableShutdownHooks()` which handles SIGTERM/SIGINT and calls `onModuleDestroy()` on all modules.

### Python (FastAPI)

```python
from contextlib import asynccontextmanager
from fastapi import FastAPI
import signal

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    logger.info("Starting application")
    await init_db()

    yield

    # Shutdown — clean up resources
    logger.info("Shutting down")
    await close_db_pool()
    logger.info("Shutdown complete")

app = FastAPI(lifespan=lifespan)

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.get("/ready")
async def ready(db: AsyncSession = Depends(get_db)):
    try:
        await db.execute(text("SELECT 1"))
        return {"status": "ok"}
    except Exception:
        return JSONResponse(status_code=503, content={"status": "error"})
```

## When to Bend the Rules

- **First deploy**: There's no old version to be compatible with. Ship the initial schema and code together.
- **Breaking changes with no traffic**: If the service has zero users (internal tool, pre-launch), skip expand-contract.
- **Emergency hotfixes**: A critical security fix can skip the expand phase if the contract phase is deferred. Document the debt.
- **Dev/staging environments**: Drop and recreate databases freely. Reserve expand-contract for production.
