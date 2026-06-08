# /review-tests

Automated test quality review skill for Claude Code. Detects test smells, rotten green tests, mock abuse, contract drift, and missing coverage.

## Usage

```
/review-tests                          # review all changed test files on current branch
/review-tests backend/tests/unit/      # review specific path
```

## What it checks

| Category | What it catches |
|----------|----------------|
| Rotten green tests | Tests that pass without verifying what they claim |
| Test smells | Assertion Roulette, Eager Test, Conditional Logic, Mystery Guest, Long Test, Obscure Test |
| Mock abuse | Over-mocking that makes tests prove nothing about real behavior |
| Contract drift | Tests coupled to implementation details instead of behavior |
| Missing coverage | Auth enforcement, authorization, input validation, cross-user isolation |
| Edge cases | Boundary values, empty collections, shared mutable state |
| Fixture fragility | Order dependence, missing cleanup, scope misuse |
| Assertion quality | Loose assertions, missing assertions, asymmetric coverage |

## Background

### Test smell taxonomy

The test smell concept originates from van Deursen et al. (2001), who catalogued recurring anti-patterns in test code analogous to code smells in production code. The taxonomy has been refined over two decades of research.

The smells detected by this skill:

- **Assertion Roulette** -- multiple unrelated assertions in one test without messages. When one fails, diagnosing which assertion broke requires reading the full traceback. (van Deursen et al., 2001)
- **Eager Test** -- a single test method exercises multiple production methods or behaviors. Violates the "test one thing" principle and makes failures ambiguous. (van Deursen et al., 2001)
- **Conditional Test Logic** -- `if/else/for/while` in test code. Tests should be linear: setup, act, assert. Branching means the test may not execute its assertions at all. (Meszaros, 2007)
- **Mystery Guest** -- test depends on external state (files, environment variables, DB rows) not visible in the test setup. Makes tests non-deterministic and hard to run in isolation. (van Deursen et al., 2001)
- **Rotten Green Test** -- test passes but doesn't verify meaningful behavior. Includes tautological assertions (`assert mock.return_value == mock.return_value`) and conditional assertions that skip silently on failure. (Delplanque et al., 2019)
- **Obscure Test** -- test name doesn't describe the behavior being verified. Forces readers to read the full body to understand intent. (Meszaros, 2007)

### LLM detection effectiveness

Santana Jr et al. (2025) evaluated LLMs on test smell detection across Java projects and found detection rates up to 96% for classic smells. LLMs outperform rule-based tools on smells that require semantic understanding (like Eager Test and Assertion Roulette) but perform comparably on structural smells (like Long Test).

Key finding: LLMs are particularly good at detecting rotten green tests because they can reason about whether an assertion actually validates the behavior described in the test name -- something static analysis tools cannot do.

### AI blind spot: shared assumptions

When AI writes both production code and tests, it carries the same assumptions into both. A bug in the mental model produces code that's wrong AND a test that validates the wrong behavior. This skill addresses this by requiring the reviewer to read production code alongside tests to check whether the test actually exercises the real risk in the code it covers.

This pattern is documented in the AI Regression Testing skill (affaan-m/everything-claude-code) which identifies four common AI testing failure modes: sandbox/production path mismatch, SELECT clause omission, error state leakage, and optimistic update failures.

### Contract drift

Tests that verify implementation details instead of observable behavior break on refactoring even when the system's behavior is unchanged. This concept comes from the "test behavior, not implementation" principle (Freeman & Pryce, 2009) and is a primary cause of brittle test suites that slow down development instead of enabling it.

### Specific mistakes LLMs make when generating tests

Empirical studies reveal systematic defects in LLM-generated tests. This skill is designed to catch these patterns specifically.

#### 1. Hallucinated APIs (34-62% of generated tests don't compile)

LLMs reference methods, classes, and parameters that don't exist. An empirical study on Defects4J found 30.68% of invalid tests fail due to "semantically-coherent but undefined identifiers" -- names that look plausible but are fabricated. Another 17.25% fail from wrong parameter types/counts, and 10.38% from attempting to instantiate abstract classes. (Chu et al., 2025; Yang et al., 2024)

#### 2. Missing triggering inputs (the #1 weakness)

74.99% of undetected defects fail because tests lack the specific input values needed to trigger bugs. Tests execute the faulty code path but use generic values that don't expose the defect. Example: triggering a complex-number bug requires setting the real part to NaN -- LLMs never generate that input. They pick "normal" values that exercise the happy path of even buggy code. (Yang et al., 2024)

#### 3. Tautological assertions

- Assert return type instead of return value (`assert isinstance(result, dict)` when the content matters)
- Mirror implementation logic in the assertion -- if the code is wrong, the test is wrong the same way
- Assert `== True` redundantly or check `'name' in data` without validating the actual value
- Check surface effects (success banner appears) without verifying real outcomes (order processed, inventory decremented)

#### 4. Near-zero mutation scores

Despite passing, LLM-generated tests approach zero mutation score because they test "ineffective logic such as interfaces or empty methods." The best LLM detected only 8 of 163 bugs on Defects4J with 0.74% precision -- roughly 135 false alarms per real fault found. (Chu et al., 2025)

#### 5. Systematic blind spots

LLM tests cluster on happy paths and consistently miss:

- **Permission boundaries** and access control edge cases
- **State transitions** and timing/concurrency bugs
- **Partial failures** (half the system fails, half continues)
- **Error branches** -- branch coverage often around 40%, skipping error handling paths entirely
- **Exception handling**: tests omit `pytest.raises` and silently accept errors as passing

#### 6. Implicit integration tests

LLMs skip mocking external dependencies, turning unit tests into integration tests that hit real APIs, fail without network, and flake randomly. The test structure looks correct but tests network availability, not business logic.

#### 7. False confidence amplification

Test generation tools often filter out failing tests before reporting, actively masking real defects. Teams treat 400 AI-generated test cases with less scrutiny than 40 hand-written ones. Green CI dashboards from AI-generated suites discourage further verification. (CodeRabbit, 2025)

## References

- van Deursen, A., Moonen, L., van den Bergh, A., & Kok, G. (2001). "Refactoring Test Code." *Proc. 2nd Int. Conf. on Extreme Programming (XP2001)*.
- Meszaros, G. (2007). *xUnit Test Patterns: Refactoring Test Code*. Addison-Wesley.
- Delplanque, J., Ducasse, S., Fuhrman, G., & Anquetil, N. (2019). "Rotten Green Tests." *Proc. 41st Int. Conf. on Software Engineering (ICSE)*.
- Freeman, S. & Pryce, N. (2009). *Growing Object-Oriented Software, Guided by Tests*. Addison-Wesley.
- Santana Jr, E.G. et al. (2025). "Evaluating LLMs Effectiveness in Detecting and Correcting Test Smells: An Empirical Study." arXiv:2506.07594.
- affaan-m/everything-claude-code (2025). "AI Regression Testing" SKILL.md. GitHub.
- Spadini, D. et al. (2020). "Investigating Severity Thresholds for Test Smells." *Proc. 17th Int. Conf. on Mining Software Repositories (MSR)*.
- Yang, L. et al. (2024). "On the Evaluation of Large Language Models in Unit Test Generation." arXiv:2406.18181. (arXiv preprint, not peer-reviewed)
- Chu, B. et al. (2025). "Large Language Models for Unit Test Generation: Achievements, Challenges, and Opportunities." arXiv:2511.21382.
- CodeRabbit (2025). "State of AI vs Human Code Generation Report." coderabbit.ai/blog.
