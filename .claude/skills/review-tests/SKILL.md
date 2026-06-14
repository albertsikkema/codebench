---
name: review-tests
description: "CB - Review test quality for changed or specified test files. Detect test smells, skip-to-pass shortcuts, mock abuse, missing coverage, and contract drift."
user-invocable: true
---

Review test quality for changed or specified test files. Detect test smells, find tests that pass without proving anything, and identify missing coverage.

## Arguments: $ARGUMENTS

## Scope

If argument is provided, review those files. Otherwise, review all test files changed on the current branch vs main.

To find changed test files when no argument given:

```bash
git diff --name-only main...HEAD -- '**/test_*' '**/test/**' '**/*_test.*' '**/*.test.*' '**/*_spec.*' '**/*.spec.*'
```

Also include untracked test files:

```bash
git ls-files --others --exclude-standard -- '**/test_*' '**/test/**' '**/*_test.*' '**/*.test.*' '**/*_spec.*' '**/*.spec.*'
```

## Critical rule

For every test file, read the test AND the production code it tests. You need both to judge whether the test actually verifies behavior. A test that looks reasonable in isolation may be testing the wrong thing, testing an outdated interface, or missing the actual risk in the code it covers.

## What to check

### 1. Rotten green tests (skip-to-pass)

Tests that pass without asserting the thing they claim to test.

- **Conditional assertions**: `if status == 200: assert ...` -- passes silently on failure
- **Bare skip**: `pytest.skip()` that hides broken setup instead of failing
- **Swallowed errors**: `try/except` that catches assertion errors
- **Tautologies**: asserting what the mock was told to return -- proves nothing about production code
- **Weak guards**: `assert True`, `assert resp is not None`, `assert len(x) >= 0` when the test name promises real verification

### 2. Test smells

Classic test smells from the literature. Flag when found:

- **Assertion Roulette**: multiple assertions without messages -- when one fails, you can't tell which without reading the traceback
- **Eager Test**: one test exercises multiple unrelated behaviors. Should be split
- **Conditional Test Logic**: `if/else/for/while` in test code. Tests should be linear
- **Mystery Guest**: depends on external state (files, env vars, DB rows) not visible in the test setup
- **Long Test**: >40 lines of test body (not counting fixtures). Usually testing too much
- **Obscure Test**: test name doesn't describe what's being verified. `test_foo_1`, `test_it_works`

### 3. Mock abuse

Tests where mocking defeats the purpose.

- `mock.assert_called_once()` with no check on arguments or result
- `session.execute.assert_called_once()` without verifying the query content
- Mocking the function under test itself
- Mock return value flows straight to assertion with no production logic in between
- Over-mocking: >3 mocks in a single test usually means the test proves nothing about real behavior

### 4. Contract drift

Tests that verify implementation details instead of behavior.

- Asserting internal method call order instead of observable output
- Testing private/internal APIs that could change without affecting behavior
- Hardcoded response structures that duplicate the implementation rather than spec
- Tests that break on refactor even when behavior is unchanged

### 5. Missing negative paths

- Auth enforcement: every endpoint that requires auth should have a 401 test
- Authorization: every endpoint with role checks should have a 403 test
- Input validation: malformed/missing fields should return 4xx
- Cross-tenant/cross-user isolation: user A cannot access user B's resources
- Error responses: verify error structure, not just status code

### 6. Missing edge cases

- Empty collections (zero results)
- Boundary values (exact timeout, exact limit -- not just "above" and "below")
- Concurrent/ordering issues (test order dependence, shared mutable state)
- Cleanup: tests that mutate shared state (like admin password) without restoring it

### 7. Fixture fragility

- Fixtures that assume specific DB state from other tests
- Module-scoped fixtures shared across tests that mutate them
- Missing cleanup that causes cascading failures
- Fixtures that do too much -- setup + action + partial assertion

### 8. Assertion quality

- Loose assertions: `assert len(x) >= 1` when exact count is known
- String matching: `assert "error" in resp.text` instead of checking structured response
- Missing assertions: test does setup + action but never asserts the result
- Asymmetric coverage: happy path has 10 assertions, error path has 1

### 9. LLM-generated test defects

Watch for these patterns especially in AI-generated or AI-assisted tests:

- **Tautological assertions**: test mirrors the implementation logic in the assertion. If the code is wrong, the test is wrong the same way. Check: does the assertion encode independent knowledge of the expected result, or does it just recompute what the code does?
- **Generic inputs only**: tests use "normal" values (positive ints, simple strings, valid emails) that don't trigger edge cases. Look for: missing NaN, None, empty string, zero, negative, boundary values, unicode, very long strings
- **Type-not-value assertions**: `assert isinstance(result, dict)` or `assert result is not None` when the actual content matters. The test name says "returns user data" but the assertion only checks it's a dict
- **Surface-level verification**: checks that a success message appears without verifying side effects (DB write, email sent, session created, inventory decremented)
- **Happy-path clustering**: all tests exercise the success path. Error branches, permission checks, and failure modes are missing entirely (~40% branch coverage is typical for LLM-generated tests)
- **Implicit integration**: unit tests that skip mocking external dependencies (DB, API, filesystem), making them flaky integration tests in disguise
- **Hallucinated APIs**: test references methods, parameters, or classes that don't exist in the codebase. Names look plausible but are fabricated

## Output format

Use the caveman review format. One line per finding.

```
<file>:L<line>: <severity>: <problem>. <fix>.
```

Severity:
- `bug:` test passes but doesn't test what it claims (rotten green)
- `risk:` test works but is fragile or could mask regressions
- `smell:` classic test smell -- not broken but degrades maintainability
- `gap:` missing test for important behavior
- `nit:` style or readability, author can ignore

## Report structure

```markdown
## Test Quality Review

### Verdict
<2-3 sentence overall assessment. State the ratio of real-behavior tests vs mock-wiring tests.>

### Findings
<one-line-per-finding, grouped by file>

### Missing Coverage
<bullet list of untested behaviors that should have tests>

### Summary Table

| Area | Verdict |
|------|---------|
| Rotten green tests | ... |
| Test smells | ... |
| Mock abuse | ... |
| Contract drift | ... |
| Auth/authz coverage | ... |
| Cross-user isolation | ... |
| Negative path coverage | ... |
| Edge cases | ... |
| LLM-generated defects | ... |
```

## Saving the review

After completing the review:

1. Run `.claude/helpers/get_metadata.sh` to collect metadata (date, branch, commit, repo) for the report header
2. Save the markdown review to `.claude/workspace/memories/` with naming: `YYYY-MM-DD-test-quality-description.md`
3. Omit sections that have no findings
