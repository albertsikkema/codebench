# Defense-in-Depth Input Validation

## Principle

Validate user input at multiple independent layers. Each layer checks different aspects of the input, so that removing or bypassing one layer does not compromise security. Validate early, validate often, and validate at every trust boundary.

## Why

- **Single-layer validation fails**: A character whitelist that allows `/` and `.` still permits `../../etc/passwd`. A regex that blocks `..` still permits URL-encoded variants. No single check catches everything.
- **Code changes over time**: A downstream function that sanitises input today may be refactored tomorrow. If upstream validation was the only line of defense, the refactor silently introduces a vulnerability.
- **Defense-in-depth is cheap**: The performance cost of multiple validation passes on typical input sizes (kilobytes) is negligible. The cost of a missed injection is not.

## The Layers

### Layer 1: Type System and Structural Constraints

Use the language's type system and framework-level validation to enforce shape, presence, and bounds.

**What it catches**: Wrong types, missing required fields, out-of-range values, empty collections, oversized payloads.

Examples:
- A list must have 1–50 items
- A string must be 1–255 characters
- A number must be positive
- A field is required, not optional

This layer is automatic and declarative. It runs before any custom logic.

### Layer 2: Format and Character Restrictions

Whitelist allowed characters or patterns. Reject input that doesn't match the expected format.

**What it catches**: Most injection characters, unexpected encoding, control characters, binary content in text fields.

The key decision is **whitelist vs blacklist**:
- **Whitelist** (preferred): Only allow known-safe characters. `[a-zA-Z0-9._-]` for identifiers.
- **Blacklist** (fragile): Block known-bad characters. Always incomplete — attackers find what you missed.

Be deliberate about which special characters you allow and why. If you allow `/` (for scoped names like `@org/package`), document the reason — and add Layer 3 checks to handle the risk it introduces.

### Layer 3: Explicit Security Pattern Checks

Check for specific attack patterns that may pass through the character whitelist. Name the attack in the code.

**What it catches**: Path traversal (`..`), absolute paths (leading `/`), null bytes, command injection patterns, SQL keywords in contexts where parameterised queries aren't possible.

These checks are explicit and documented:

```
// Prevent path traversal — the character whitelist allows "." and "/"
// individually, but ".." as a sequence is never legitimate input.
if strings.Contains(input, "..") {
    return ErrPathTraversal
}
```

The comment explains *why* this check exists even though Layers 1 and 2 are already in place. This prevents a future maintainer from removing it as "redundant."

### Layer 4: Downstream Sanitisation

Before using validated input in a sensitive context (filename, query parameter, shell argument), apply a final transformation appropriate to that context.

**What it catches**: Edge cases that slipped through, encoding issues, platform-specific quirks.

Examples:
- Replace non-alphanumeric characters with `_` when generating filenames
- Use parameterised queries for database operations (not string interpolation)
- Shell-escape arguments before passing to `exec`
- HTML-encode before inserting into templates

This layer is the last line of defense. It should be simple, mechanical, and context-specific.

## Independence Between Layers

Each layer must provide security value on its own. Test this mentally:

- "If I remove Layer 2 (character whitelist), does Layer 3 still catch path traversal?" → Yes.
- "If I remove Layer 3, does Layer 2 still block most injection attempts?" → Yes.
- "If both are removed, does Layer 4 (parameterised queries) still prevent SQL injection?" → Yes.

If removing one layer defeats all security, your layers are not independent — they are a single layer split across multiple functions.

## Error Messages

Security validation errors should be specific enough to help legitimate users fix mistakes, but should not reveal internal implementation details.

```
GOOD: "Package name contains '..' which is not allowed"
GOOD: "Input contains characters outside the allowed set [a-zA-Z0-9._-]"
BAD:  "Invalid input"           (not actionable)
BAD:  "SQL injection detected"  (reveals detection mechanism)
```

## Validation Order

Validate from general to specific, cheap to expensive:

1. **Type and presence** (framework/type system — free)
2. **Format and characters** (string scan — microseconds)
3. **Security patterns** (pattern match — microseconds)
4. **Business rules** (may require database lookup — milliseconds)

Reject as early as possible. Don't hit the database to check a foreign key if the input contains path traversal characters.

## Implementation Notes

### Go

Go's strong typing handles Layer 1 naturally — struct fields with types, required vs pointer fields. For Layers 2–3, validate in a dedicated `Validate() error` method on request types.

```go
type CreateRequest struct {
    Name string `json:"name" validate:"required,min=1,max=255"`
    Items []string `json:"items" validate:"required,min=1,max=50,dive,min=1"`
}

func (r *CreateRequest) Validate() error {
    for i, item := range r.Items {
        // Layer 2: character whitelist
        if !safeNamePattern.MatchString(item) {
            return fmt.Errorf("item %d contains invalid characters", i)
        }
        // Layer 3: explicit security checks
        if strings.Contains(item, "..") || strings.HasPrefix(item, "/") {
            return fmt.Errorf("item %d contains path traversal pattern", i)
        }
    }
    return nil
}
```

Use libraries like `go-playground/validator` for struct tag validation, but don't rely on it exclusively — add explicit checks for security patterns.

### TypeScript

Use [Zod](https://zod.dev/) for runtime schema validation. Zod gives you Layer 1 (types and constraints) and integrates cleanly with custom refinements for Layers 2–3. The parsed output is fully typed — downstream code can trust it.

```typescript
import { z } from "zod";

const ResourceRequest = z.object({
  // Layer 1: type, presence, bounds
  items: z
    .array(z.string().min(1).max(255))
    .min(1)
    .max(50),
}).transform((data) => ({
  ...data,
  items: data.items.map((item) => {
    const trimmed = item.trim();

    // Layer 2: character whitelist
    if (!/^[a-zA-Z0-9\-_@/.]+$/.test(trimmed)) {
      throw new Error(`'${trimmed}' contains invalid characters`);
    }

    // Layer 3: security patterns
    if (trimmed.includes("..") || trimmed.startsWith("/") || trimmed.endsWith("/")) {
      throw new Error(`'${trimmed}' contains path traversal pattern`);
    }

    return trimmed;
  }),
}));

type ResourceRequest = z.infer<typeof ResourceRequest>;
```

For Express/Fastify, validate in middleware or at the handler boundary. For NestJS, use Zod with `@anatine/zod-nestjs` or the built-in `ValidationPipe` with `class-validator`.

### Python (Pydantic)

Use [Pydantic](https://docs.pydantic.dev/) models for declarative, multi-layer validation. `Field()` constraints handle Layer 1, `@field_validator` decorators handle Layers 2–3, and Pydantic's automatic type coercion and error reporting give you clear, structured validation errors for free.

```python
class ResourceRequest(BaseModel):
    # Layer 1: type, presence, bounds
    items: list[str] = Field(..., min_length=1, max_length=50)

    @field_validator("items")
    @classmethod
    def validate_items(cls, v: list[str]) -> list[str]:
        cleaned = []
        for item in v:
            item = item.strip()
            if not item:
                raise ValueError("Item cannot be empty")

            # Layer 2: character whitelist
            if not all(c.isalnum() or c in "-_@/." for c in item):
                raise ValueError(f"'{item}' contains invalid characters")

            # Layer 3: security patterns
            if ".." in item or item.startswith("/") or item.endswith("/"):
                raise ValueError(f"'{item}' contains path traversal pattern")

            cleaned.append(item)
        return cleaned
```

## When to Bend the Rules

- **Internal APIs between trusted services**: Layer 1 (types) is usually sufficient. Layers 2–3 are for trust boundaries with user input.
- **Read-only display of user input**: If input is only displayed (never used as a filename, query, or command), Layer 4 sanitisation can be HTML-encoding only.
- **Performance-critical hot paths**: If you process millions of items per second, consider validating once at ingestion and trusting the data downstream. But measure before optimising — validation is rarely the bottleneck.

The default should always be to validate. Skipping a layer is a conscious decision that should be documented with the reason.
