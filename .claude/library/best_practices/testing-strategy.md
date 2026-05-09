# Testing Strategy

## Principle

Test behavior, not implementation. Organise tests into tiers with different speed, scope, and cost trade-offs. Every feature tests both the happy path and the relevant error paths. A test suite that takes minutes to run will eventually stop being run.

## Why

- **Fast feedback changes behavior**: When tests run in 2 seconds, developers run them constantly. When tests take 10 minutes, developers push and hope. Test speed directly affects code quality.
- **Tests are documentation**: A well-written test says "given these inputs, the system produces these outputs." When the test name reads like a requirement, the test suite becomes a living specification.
- **Testing implementation creates brittle tests**: Tests that verify internal method calls break when code is refactored, even when behavior is preserved. Tests that verify inputs and outputs survive refactoring.

## The Test Pyramid

```
        /  E2E / LLM  \        ← Few, slow, expensive
       / Integration    \       ← Moderate, real dependencies
      /   Unit           \      ← Many, fast, isolated
     ‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
```

### Tier 1: Unit Tests

**Speed**: Seconds (entire suite < 5s)
**Scope**: Single function, method, or module in isolation
**Dependencies**: All external dependencies mocked or stubbed
**Runs**: On every save, every commit, every CI pipeline

**What to test**:
- Business logic and domain rules
- Data transformations and calculations
- Input validation and edge cases
- Error handling paths
- State machines and conditional logic

**What not to test**:
- Trivial getters/setters with no logic
- Framework boilerplate (the framework is already tested)
- Private implementation details that may change
- Exact log messages or output formatting

### Tier 2: Integration Tests

**Speed**: Seconds to low minutes (suite < 60s)
**Scope**: Multiple components working together, with real infrastructure
**Dependencies**: Real database (Docker), real filesystem; external APIs mocked
**Runs**: Before merge, in CI

**What to test**:
- API endpoints end-to-end (request → response)
- Database operations (queries, transactions, constraints)
- Authentication and authorization flows
- Middleware behavior (CORS, rate limiting, error handling)

**What to mock**:
- External third-party APIs (payment gateways, email services)
- Services that cost money per call (LLM APIs, SMS)
- Services that are flaky or slow in test environments

### Tier 3: End-to-End / System Tests

**Speed**: Minutes
**Scope**: Full system, often including UI
**Dependencies**: Everything real (or close to it)
**Runs**: Before release, manually, or on a schedule

**What to test**:
- Critical user flows (signup, purchase, core workflow)
- Integration with real external services (when cost-justified)
- Performance under load (separate from functional tests)

**What not to test here**:
- Edge cases (that's what unit tests are for)
- Every permutation of inputs (combinatorial explosion)

## Core Rules

### 1. Test Behavior, Not Implementation

```
BAD:  Assert that the service called repository.findAll() exactly once
      → Breaks when you add caching, change the query method, or refactor internals

GOOD: Assert that given these inputs, the API returns these outputs
      → Survives refactoring, catches real regressions
```

### 2. Every Feature Tests Happy Path and Error Paths

For each feature, test:
- **Happy path**: Expected inputs produce expected outputs
- **Validation errors**: Invalid input is rejected with appropriate error
- **Auth errors**: Unauthenticated/unauthorized requests are rejected
- **Downstream failures**: What happens when a dependency fails
- **Edge cases**: Empty collections, maximum values, concurrent access

### 3. Mock at System Boundaries, Not Internally

```
BAD:  Mock the internal UserService when testing the UserHandler
      → Tests pass even if UserService has bugs

GOOD: Mock the database (or use a test database) when testing the UserHandler
      → Handler + Service + Repository are tested together, only the DB is replaced
```

The further out you push your mock boundary, the more real code you test. Push it to the infrastructure edge.

### 4. Tests Must Be Independent and Repeatable

- Each test sets up its own state and cleans up after itself
- Tests can run in any order and still pass
- Tests can run in parallel without interfering with each other
- Running the same test twice produces the same result

For database tests: use transactions that roll back after each test, or truncate tables in a fixture.

### 5. Use Test Fixtures, Not Copy-Paste Setup

```
BAD:  Every test file creates its own User, Order, and Session objects
      with slightly different field values

GOOD: A shared fixture or factory provides standard test objects
      Tests override only the fields relevant to what they're testing
```

### 6. Name Tests as Specifications

```
BAD:  test_user_1, test_create, test_error

GOOD: test_create_user_with_valid_email_returns_201
      test_create_user_with_duplicate_email_returns_409
      test_delete_order_by_non_owner_returns_404
```

The test name should describe the scenario and expected outcome. When a test fails, the name alone should tell you what's broken.

## Coverage

**Target**: 80% overall, 70% minimum per module.

Coverage measures which lines execute during tests, not whether the tests are meaningful. High coverage with bad assertions is worse than moderate coverage with good assertions.

**What to prioritise for coverage**:
- Business logic: aim for 90%+
- API handlers: aim for 80%+
- Data access: aim for 70%+ (integration tests cover this)
- Configuration/setup: don't chase coverage here

## Test Framework Bootstrap

Every new project should include, from day one:
1. A configured test framework with a `tests/` directory
2. An example unit test for a pure function
3. An example integration test with fixture setup/teardown
4. An example of mocking an external dependency

These examples are the project's test conventions — new tests are written by copying and adapting them.

## Implementation Notes

### Go

```go
// Unit test — table-driven
func TestCalculateTotal(t *testing.T) {
    tests := []struct {
        name     string
        items    []Item
        expected int64
    }{
        {"empty cart", nil, 0},
        {"single item", []Item{{Price: 1000, Qty: 2}}, 2000},
        {"multiple items", []Item{{Price: 500, Qty: 1}, {Price: 300, Qty: 3}}, 1400},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculateTotal(tt.items)
            if got != tt.expected {
                t.Errorf("got %d, want %d", got, tt.expected)
            }
        })
    }
}

// Integration test — use testcontainers or docker-compose
func TestCreateUser_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    db := setupTestDB(t) // start postgres in container
    repo := NewUserRepo(db)

    user, err := repo.Create(ctx, &User{Email: "test@example.com"})
    require.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

Use `go test -short` to skip integration tests during development. Use `testify` for assertions or stick with the standard library.

### TypeScript (Vitest / Jest)

```typescript
import { describe, it, expect } from "vitest";

// Unit test — parameterized
describe("calculateTotal", () => {
  it.each([
    { items: [], expected: 0 },
    { items: [{ price: 1000, qty: 2 }], expected: 2000 },
    { items: [{ price: 500, qty: 1 }, { price: 300, qty: 3 }], expected: 1400 },
  ])("returns $expected for $items.length items", ({ items, expected }) => {
    expect(calculateTotal(items)).toBe(expected);
  });
});

// Integration test with supertest
import request from "supertest";
import { app } from "../src/app";

describe("POST /users", () => {
  it("creates a user and returns 201", async () => {
    const res = await request(app)
      .post("/users")
      .send({ email: "test@example.com", name: "Test" });

    expect(res.status).toBe(201);
    expect(res.body.id).toBeDefined();
  });

  it("returns 409 for duplicate email", async () => {
    await request(app).post("/users").send({ email: "dup@example.com", name: "First" });
    const res = await request(app).post("/users").send({ email: "dup@example.com", name: "Second" });

    expect(res.status).toBe(409);
  });
});
```

Use Vitest (faster, ESM-native) or Jest. Use `supertest` for HTTP integration tests. For database tests, use transactions that roll back in `afterEach`, or `testcontainers` for PostgreSQL.

### Python

```python
import pytest

# Unit test with parametrize
@pytest.mark.parametrize("items,expected", [
    ([], 0),
    ([{"price": 1000, "qty": 2}], 2000),
    ([{"price": 500, "qty": 1}, {"price": 300, "qty": 3}], 1400),
])
def test_calculate_total(items, expected):
    assert calculate_total(items) == expected

# Integration test with real database
@pytest.mark.integration
@pytest.mark.asyncio
async def test_create_user(db_session):
    repo = UserRepo(db_session)
    user = await repo.create(email="test@example.com")
    assert user.id is not None

# Testing error paths
@pytest.mark.asyncio
async def test_create_user_duplicate_email_raises(db_session):
    repo = UserRepo(db_session)
    await repo.create(email="test@example.com")
    with pytest.raises(DuplicateEmailError):
        await repo.create(email="test@example.com")
```

Use `pytest` markers (`@pytest.mark.unit`, `@pytest.mark.integration`) to separate tiers. Run unit tests with `pytest -m unit` for fast feedback.

## When to Bend the Rules

- **Prototypes**: Skip integration tests. Write enough unit tests to validate the core logic. Add integration tests before production.
- **Stable CRUD endpoints**: If a handler is pure delegation (validate → call service → return), a single integration test per endpoint is sufficient. Don't write unit tests for trivial pass-through code.
- **Legacy code without tests**: Don't try to add unit tests to tightly coupled legacy code. Start with integration tests that verify external behavior, then refactor toward testability.
