# Background Job Patterns

## Principle

Not all work belongs in the request/response cycle. Operations that are slow, unreliable, or don't need an immediate result should run in the background. Choose the right pattern — fire-and-forget, tracked jobs, or queue-based processing — based on how much visibility and control the caller needs.

## Why

- **Users shouldn't wait for slow operations**: Sending an email, generating a report, processing an upload — these can take seconds to minutes. Making the user wait blocks them and wastes server resources holding the connection open.
- **Reliability requires decoupling**: If a payment notification depends on an email service that's temporarily down, the payment shouldn't fail. Background processing with retries decouples the critical path from non-critical dependencies.
- **Resource management**: CPU-intensive tasks (image processing, PDF generation, data aggregation) should run outside the request handler's thread/goroutine pool to avoid starving incoming requests.

## The Three Patterns

### 1. Fire-and-Forget

Return immediately. Spawn background work. Don't track status.

```
Client → POST /reports/generate → 202 Accepted
                                    ↓
                              Background worker
                                    ↓
                              Results appear in filesystem/email/S3
```

**Use when**:
- The caller doesn't need progress updates
- Results are available out-of-band (filesystem, email, object storage)
- Failure can be logged but doesn't need immediate user notification
- The operation is self-contained (no cancellation or resumption needed)

**Implementation**:

```
HTTP Handler (synchronous):
  1. Validate input (fail fast if invalid)
  2. Prepare resources (create output directories, validate permissions)
  3. Spawn background worker (daemon thread / goroutine / detached task)
  4. Return 202 Accepted immediately

Background Worker:
  1. Execute the operation
  2. Log success or failure with full context
  3. Write results to the agreed-upon location
```

**Key rules**:
- Validate and prepare resources in the handler, before spawning — fail fast on bad input
- Use daemon threads/goroutines so they don't prevent process shutdown
- Log with a correlation ID so you can trace the work back to the originating request
- Include thread/task identifiers in logs for concurrent job debugging

### 2. Tracked Jobs

Return immediately with a job ID. Track status in a database. Client polls or receives a callback.

```
Client → POST /exports → 202 Accepted { "job_id": "abc-123" }
                            ↓
                      Background worker
                      Updates status: PENDING → RUNNING → COMPLETED
                            ↓
Client → GET /jobs/abc-123 → { "status": "COMPLETED", "result_url": "..." }
```

**Use when**:
- Users need to see progress or status
- The operation can fail and the user needs to know
- Cancellation or resumption might be needed
- Results are delivered through the API, not out-of-band

**Status lifecycle**:

```
PENDING → RUNNING → COMPLETED
                  → FAILED (with error message)
          ↑
       CANCELLED (if cancellation is supported)
```

**Implementation**:

```
HTTP Handler:
  1. Validate input
  2. Create job record in database (status: PENDING)
  3. Enqueue or spawn background worker with job ID
  4. Return 202 with job ID

Background Worker:
  1. Update status to RUNNING
  2. Execute operation (with progress updates if applicable)
  3. On success: update status to COMPLETED, store result
  4. On failure: update status to FAILED, store error details

Status Endpoint:
  GET /jobs/{id} → returns current status, progress, result/error
```

**Key rules**:
- Store the full error message in the job record for debugging (not just "failed")
- Set a timeout — jobs that run forever should be marked FAILED after a deadline
- Make the status endpoint cheap (database read only, no computation)

### 3. Queue-Based Processing

Decouple producer and consumer with a message queue. The most robust pattern for high-volume or distributed systems.

```
Producer → Message Queue → Consumer(s)
                ↓
          Retry / Dead Letter Queue
```

**Use when**:
- Multiple producers and/or multiple consumers
- Work needs to survive process restarts
- You need guaranteed delivery (at-least-once processing)
- Load varies and you need to absorb spikes
- Work should be distributed across multiple workers

**Components**:

| Component | Responsibility |
|-----------|---------------|
| **Producer** | Validates input, publishes message to queue |
| **Queue** | Stores messages durably, delivers to consumers |
| **Consumer** | Processes one message at a time, acknowledges on completion |
| **Dead Letter Queue (DLQ)** | Stores messages that failed after max retries |

**Key rules**:
- **Idempotent consumers**: Messages may be delivered more than once. Processing the same message twice must produce the same result.
- **Acknowledge after completion**: Don't ack the message before the work is done. If the consumer crashes, the message should be redelivered.
- **Bounded retries**: Set a max retry count. After N failures, move to the DLQ for manual investigation.
- **Monitor queue depth**: Growing queue depth means consumers can't keep up — alert on this.

## Cross-Cutting Concerns

### Isolation

Background workers must not share mutable state with the request handler.

- **Go**: Goroutines are naturally isolated. Pass data by value or through channels.
- **TypeScript**: Use worker threads for CPU-intensive work. For I/O work, `setImmediate` or `setTimeout(fn, 0)` is sufficient.
- **Python**: Use separate threads with their own event loops for async work. The GIL limits CPU parallelism — use `multiprocessing` or a task queue (Celery, Dramatiq) for CPU-bound work.

### Logging

Background jobs run without a request context. Create your own:

- Generate a job ID or correlation ID at spawn time
- Include it in every log line from the worker
- Include thread/goroutine/task identifiers for concurrent debugging
- Log start, completion, and failure with timing information

```
[job:abc-123] Starting report generation (3 items)
[job:abc-123] Processing item 1/3: success
[job:abc-123] Processing item 2/3: failed (timeout after 30s)
[job:abc-123] Processing item 3/3: success
[job:abc-123] Completed: 2/3 succeeded, 1/3 failed (total: 45s)
```

### Timeouts

Every background job needs a maximum runtime. Without it, a hung job consumes resources forever.

- Set a deadline at spawn time
- Check the deadline periodically during long operations
- Kill or mark as FAILED when the deadline expires
- Make the timeout configurable per job type

### Graceful Shutdown

When the process receives SIGTERM:
- Stop accepting new jobs
- Let in-progress jobs finish (with a timeout)
- For queue-based: stop consuming, finish current message, don't ack new ones
- For fire-and-forget: daemon threads/goroutines are killed automatically — accept that in-progress work may be lost

## Implementation Notes

### Go

```go
// Fire-and-forget with goroutine
func (h *Handler) GenerateReport(w http.ResponseWriter, r *http.Request) {
    req, err := parseRequest(r)
    if err != nil {
        writeProblem(w, 400, "Invalid request", err.Error())
        return
    }

    // Prepare resources before spawning
    outputDir := filepath.Join("reports", req.ID)
    if err := os.MkdirAll(outputDir, 0o755); err != nil {
        writeProblem(w, 500, "Setup failed", "Could not create output directory")
        return
    }

    jobID := uuid.NewString()

    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
        defer cancel()

        logger := slog.With("job_id", jobID, "type", "report")
        logger.Info("starting report generation", "items", len(req.Items))

        for i, item := range req.Items {
            if err := processItem(ctx, item); err != nil {
                logger.Error("item failed", "item", item, "error", err)
                continue // graceful degradation
            }
            logger.Info("item completed", "index", i+1, "total", len(req.Items))
        }

        logger.Info("report generation complete")
    }()

    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "started",
        "job_id": jobID,
        "message": fmt.Sprintf("Report will be available in %s", outputDir),
    })
}
```

### TypeScript

```typescript
// Tracked job with BullMQ (Redis-backed queue)
import { Queue, Worker } from "bullmq";

const reportQueue = new Queue("reports", { connection: redis });

// Producer (HTTP handler)
app.post("/reports/generate", async (req, res) => {
  const validated = ReportRequest.parse(req.body);

  const job = await reportQueue.add("generate", {
    items: validated.items,
    userId: req.user.id,
  }, {
    attempts: 3,
    backoff: { type: "exponential", delay: 1000 },
    removeOnComplete: 100, // keep last 100 completed jobs
    removeOnFail: 200,
  });

  res.status(202).json({
    jobId: job.id,
    status: "pending",
    statusUrl: `/jobs/${job.id}`,
  });
});

// Consumer (worker)
const worker = new Worker("reports", async (job) => {
  const { items, userId } = job.data;
  logger.info({ jobId: job.id, items: items.length }, "starting report");

  for (const [i, item] of items.entries()) {
    await processItem(item);
    await job.updateProgress((i + 1) / items.length * 100);
  }

  return { outputPath: `/reports/${job.id}.pdf` };
}, { connection: redis, concurrency: 5 });

// Status endpoint
app.get("/jobs/:id", async (req, res) => {
  const job = await reportQueue.getJob(req.params.id);
  if (!job) return res.status(404).json({ error: "Job not found" });

  const state = await job.getState();
  res.json({
    id: job.id,
    status: state,
    progress: job.progress,
    result: job.returnvalue,
    error: job.failedReason,
  });
});
```

### Python

```python
# Fire-and-forget with threading
import threading
import asyncio

def trigger_report(items: list[str]) -> str:
    """Spawn background report generation, return immediately."""
    job_id = str(uuid.uuid4())

    def run_in_thread():
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            loop.run_until_complete(generate_report(job_id, items))
        finally:
            loop.close()

    thread = threading.Thread(target=run_in_thread, daemon=True)
    thread.start()
    return job_id

async def generate_report(job_id: str, items: list[str]) -> None:
    logger.info("Starting report", extra={"job_id": job_id, "item_count": len(items)})

    for i, item in enumerate(items):
        try:
            await process_item(item)
            logger.info(f"Item {i+1}/{len(items)} done", extra={"job_id": job_id})
        except Exception as e:
            logger.error(f"Item failed: {e}", extra={"job_id": job_id}, exc_info=True)
            continue

    logger.info("Report complete", extra={"job_id": job_id})

# FastAPI endpoint
@router.post("/reports/generate", status_code=202)
def generate(request: ReportRequest):
    job_id = trigger_report(request.items)
    return {"status": "started", "job_id": job_id}
```

For production queue-based processing in Python, use Celery (Redis/RabbitMQ), Dramatiq, or ARQ. Threading is sufficient for fire-and-forget; task queues are better for tracked and retryable jobs.

## Choosing the Right Pattern

| Question | Fire-and-forget | Tracked | Queue-based |
|----------|----------------|---------|-------------|
| Does the user need status updates? | No | Yes | Yes |
| Does failure need user notification? | No | Yes | Yes |
| Need cancellation/resumption? | No | Maybe | Yes |
| Must survive process restarts? | No | Yes (DB) | Yes (queue) |
| Multiple workers needed? | No | Maybe | Yes |
| High volume / spiky load? | No | No | Yes |

Start with fire-and-forget. Move to tracked when users need visibility. Move to queues when you need durability, distribution, or retries.

## When to Bend the Rules

- **Sub-second background work**: If the operation takes <100ms, just do it inline in the request handler. The overhead of spawning a thread or publishing to a queue isn't worth it.
- **Serverless/edge functions**: No persistent workers. Use platform-provided queues (SQS, Cloud Tasks, Vercel Cron) instead of in-process background threads.
- **Single-user tools**: Fire-and-forget with filesystem output is perfectly fine. Don't add Redis and a job table for a tool with one user.
