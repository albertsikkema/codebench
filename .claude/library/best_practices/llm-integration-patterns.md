# LLM Integration Patterns

## Principle

LLM calls are expensive, slow, and non-deterministic. Design around these constraints: route cheap before expensive, keep agents stateless when possible, format context deliberately with truncation limits, and manage prompts as versioned files. Every LLM call should have a clear purpose, a bounded cost, and observable behavior.

## Why

- **Cost compounds quickly**: A single LLM call costs fractions of a cent. A thousand users making ten requests each at 2,000 tokens per request adds up fast. Architecture decisions that seem minor — skipping a classifier, including too much context — can double your LLM spend.
- **Latency is user-facing**: An LLM call takes 500ms–3s. Users notice. Every unnecessary call, every oversized context window, every retry adds to the wait.
- **Non-determinism is the default**: The same prompt can produce different outputs on different calls. Your architecture must handle this — don't assume LLM responses are consistent, structured, or correct.

## Core Rules

### 1. Keep Deterministic Work Outside the LLM

An LLM is non-deterministic, slow, and expensive. Regular code is deterministic, fast, and free. Push as much work as possible into regular code before and after the LLM call. The LLM should only do what regular code cannot: understand intent, reason, generate natural language.

```
BAD:  "Here is a JSON blob with 50 records. Filter to the ones
       from 2025 and sort by score. Then summarize the top 5."

      → The LLM does filtering, sorting, AND summarizing.
        Filtering and sorting are deterministic — they don't
        need an LLM. You're paying tokens for work a database
        query or three lines of code can do perfectly.

GOOD: Code filters to 2025 records.
      Code sorts by score and takes top 5.
      Code formats the 5 records into a context string.
      LLM summarizes the 5 records.

      → The LLM only does the part that requires language understanding.
```

**Before the LLM call** (deterministic):
- Filter, sort, deduplicate data
- Validate and normalize input
- Look up reference data (user profile, settings, history)
- Format context with truncation limits
- Select the right prompt template

**The LLM call** (non-deterministic):
- Understand intent or classify
- Reason about ambiguous input
- Generate natural language
- Decide which tool to call (if using agents)

**After the LLM call** (deterministic):
- Parse and validate the response against a schema
- Extract structured fields (regex, JSON parse)
- Apply business rules to the result
- Format the final output
- Track citations, costs, token usage

This principle reduces cost (fewer tokens), improves reliability (deterministic code doesn't hallucinate), and makes the system easier to test (you can unit-test the pre/post processing without calling an LLM).

### 2. Stateless Agents for Pure Transforms

Not every LLM task needs tools, dependencies, or state accumulation. Identify which tasks are pure text-in, text-out transforms and keep them simple.

**Use a stateless agent when the task**:
- Takes text input and produces text output
- Doesn't need to call external services
- Doesn't accumulate state across tool calls
- Doesn't need user/session context

**Examples of stateless tasks**:
- Summarization (search results → summary)
- Classification (message → intent category)
- Translation (English → Dutch)
- Reformatting (JSON → natural language)
- Extraction (document → structured fields)

**Examples requiring state/tools**:
- Search-augmented generation (needs search tool)
- Multi-step reasoning with data retrieval (needs database access)
- Citation tracking (accumulates references across tool calls)

Stateless agents are faster, cheaper, easier to test, and have fewer failure modes. Default to stateless; add dependencies only when the task requires them.

### 3. Format Context Deliberately

When feeding data to an LLM (search results, database records, documents), don't dump raw data. Create a dedicated formatting function that produces a clean, truncated, token-efficient context string.

**Rules**:
- **Truncate individual items**: Each search result summary capped at 300 chars, each answer at 200 chars
- **Limit collection size**: Top 5 results, not all 50
- **Include structure**: Numbered items, clear labels, scores/relevance indicators
- **Calculate token budget**: 1,000 chars ≈ 250 tokens. Know your budget before formatting.

```
GOOD:
  User searched for: "contract breach"

  Top 5 Results:

  1. [Score: 0.95] Supreme Court, 12-01-2024 (ECLI:NL:HR:2024:123)
     Summary: This case concerns breach of contract in...

  2. [Score: 0.87] District Court, 05-03-2023
     Summary: The plaintiff alleged breach of...

BAD:
  {"results": [{"id": "abc", "score": 0.95, "metadata": {...}, "content": "<5000 chars>", ...}, ...]}
```

**Make the formatter a pure function**: Takes structured data in, returns a string out. No side effects, easy to test, easy to adjust truncation limits.

**For LLM output**: When you need structured data back from the LLM, force JSON output mode (most providers support this) rather than hoping the LLM returns parseable text. This moves parsing from non-deterministic (regex over free-form text) to deterministic (JSON.parse over guaranteed JSON). Always validate the parsed JSON against a schema — the structure is guaranteed, the content is not.

### 4. Manage Prompts as Files

Store system prompts in versioned files (Markdown with YAML frontmatter), not hardcoded in application code.

```
prompts/
├── intent-classifier.md
├── search-summarizer.md
└── domain-agent.md
```
**Benefits**:
- Git history tracks prompt evolution
- Non-developers can review and suggest changes
- Prompts can be loaded and cached at startup
- A/B testing different prompt versions is straightforward

**Constraint**: Tool descriptions (what the LLM sees about available tools) typically must stay in code — most frameworks require them as function docstrings or schema definitions. Document this boundary clearly.

### 5. Cache LLM Responses Where Appropriate

LLM calls are expensive and often produce similar results for similar inputs. Cache when:

- The same question is asked frequently (FAQ-style queries)
- The input is deterministic (classification of known categories)
- Freshness is not critical (summaries of static content)

**Don't cache when**:
- The response depends on real-time data
- The user expects a unique/creative response
- The input contains user-specific context

Use semantic similarity for cache keys when exact-match caching isn't sufficient — but start with exact match (it's simpler and catches more than you'd expect).

### 6. Handle Non-Determinism Explicitly

The same prompt can produce different outputs. Design for this:

- **Classification**: Parse the response strictly. If the classifier returns an unexpected value, fall back to a safe default (e.g. UNCLEAR) and log the fallback for monitoring.
- **Structured output**: Validate the response against a schema. If it doesn't parse, retry once or return an error — don't try to fix malformed output with string manipulation.
- **Monitoring**: Track fallback rates. If your classifier falls back more than 5% of the time, the prompt needs work.

### 7. Never Expose Raw LLM Errors to Users

LLM failures (rate limits, timeouts, malformed responses) should be caught and translated to user-friendly messages. Never expose:

- Model names or provider details
- Token counts or rate limit internals
- Raw error messages from the provider API
- The prompt that was sent

Log the full details server-side. Return a generic message to the user.

### 8. Observe Everything

LLM calls are the hardest part of your system to debug. Instrument them thoroughly:

- **Input**: Prompt length (tokens), model used, temperature/parameters
- **Output**: Response length, finish reason, token usage (prompt + completion)
- **Timing**: Latency per call, time-to-first-token for streaming
- **Tools**: Which tools were called, in what order, with what arguments, what they returned
- **Errors**: Rate limit hits, timeouts, malformed responses, fallback events
- **Cost**: Track per-request cost (input tokens × price + output tokens × price)

Use OpenTelemetry-compatible tracing (spans per LLM call, per tool execution) so you can see the full request lifecycle.

## Implementation Notes

### Go

```go
// Two-stage routing
func (s *ChatService) HandleMessage(ctx context.Context, msg string) (string, error) {
    // Stage 1: Classify (cheap model, no tools)
    intent, err := s.classifier.Classify(ctx, msg)
    if err != nil {
        slog.WarnContext(ctx, "classifier failed, defaulting to ON_TOPIC", "error", err)
        intent = IntentOnTopic // safe default
    }

    switch intent {
    case IntentOffTopic:
        return "I can only help with domain-specific questions.", nil
    case IntentUnclear:
        return "Could you clarify your question?", nil
    case IntentOnTopic:
        return s.agent.Run(ctx, msg) // Stage 2: expensive
    default:
        slog.WarnContext(ctx, "unknown intent, defaulting to agent", "intent", intent)
        return s.agent.Run(ctx, msg)
    }
}
```

### TypeScript

```typescript
// Context formatter — pure function with truncation
function formatSearchContext(
  results: SearchResult[],
  query: string,
  maxResults = 5,
): string {
  const top = results.slice(0, maxResults);

  let context = `User searched for: "${query}"\n\n`;
  context += `Top ${top.length} Results:\n\n`;

  for (const [i, result] of top.entries()) {
    context += `${i + 1}. [Score: ${result.score.toFixed(2)}] ${result.title}\n`;
    context += `   Summary: ${result.summary.slice(0, 300)}`;
    if (result.summary.length > 300) context += "...";
    context += "\n\n";
  }

  return context;
}

// Prompt loading with caching
const promptCache = new Map<string, string>();

function loadPrompt(name: string): string {
  if (promptCache.has(name)) return promptCache.get(name)!;

  const content = fs.readFileSync(`prompts/${name}.md`, "utf-8");
  // Strip YAML frontmatter
  const stripped = content.replace(/^---\n[\s\S]*?\n---\n/, "").trim();
  promptCache.set(name, stripped);
  return stripped;
}
```

### Python

```python
# Two-stage routing with fallback tracking
async def handle_message(self, message: str, user_id: UUID) -> str:
    # Stage 1: Classify
    fallback = False
    try:
        intent_raw = await self.classifier.run(message)
        intent = Intent(intent_raw.output.strip().upper())
    except ValueError:
        fallback = True
        logger.warning(
            "Intent fallback",
            extra={
                "metric_type": "intent_fallback",
                "raw_output": intent_raw.output[:100],
                "user_id": str(user_id),
            },
        )
        intent = Intent.UNCLEAR

    logger.info(
        f"Intent classified: {intent.value}",
        extra={
            "intent": intent.value,
            "is_fallback": fallback,
            "user_id": str(user_id),
        },
    )

    # Stage 2: Route
    if intent == Intent.OFF_TOPIC:
        return "I can only help with domain-specific questions."
    elif intent == Intent.UNCLEAR:
        return "Could you clarify your question?"
    else:
        return await self.agent.run(message, deps=self.deps)
```

## When to Bend the Rules

- **Prototypes**: Skip the classifier. Send everything to the full agent. Add routing when cost matters.
- **Low-volume internal tools**: The cost optimization of two-stage routing may not justify the complexity if you have 50 users making 10 requests per day.
- **Creative applications** (writing assistants, brainstorming tools): Non-determinism is a feature, not a bug. Don't over-constrain the output.
- **Simple single-turn tasks**: If every request legitimately needs the full agent, the classifier adds latency without saving cost. Measure before adding it.
