# Engineering Principles

## KISS — Keep It Simple

- Choose the simplest solution that works
- Avoid abstractions until they're clearly needed
- Prefer readable, straightforward code over clever code
- Simple is better than complicated — simple is elegant
- Keep the codebase simple — most of what you think is necessary actually isn't
- If a solution needs a long explanation, it's probably too complex

## Fight Bloat

- Actively refactor and simplify — less code is better code
- Reduce lines of code wherever possible without losing clarity
- Question the project structure — if it's overengineered, flatten it
- Remove files, abstractions, and indirection that don't earn their keep

## YAGNI — You Ain't Gonna Need It

- Only build what's needed right now
- Don't add features, parameters, or abstractions for hypothetical future use
- Delete dead code — don't comment it out "just in case"
- One concrete use case beats three speculative ones

## Always Test

- Run existing tests after every change to catch regressions immediately
- Add tests for new functionality before considering it done
- If a bug is fixed, add a test that would have caught it
- Don't trust code that hasn't been tested
- Don't over-mock — if everything is mocked, the test proves nothing
- Tests should exercise real behavior, not just verify mock wiring

## Work in Steps

- Make one logical change at a time
- Verify each step works before moving to the next
- Commit after each working step — not in one big batch at the end
- If something breaks, the last step is easy to identify and revert
- Don't try to implement all requirements/specs in one go — spread them across multiple steps
- Each step should deliver a working, testable increment

## Fail Fast

- Surface errors early and loudly — don't silently continue with bad state
- Don't swallow exceptions or return defaults when something is actually wrong

## Single Responsibility

- Each function, module, and file does one thing
- If you can't describe what it does in one sentence, split it

## Boy Scout Rule

- Leave code better than you found it
- Small cleanup while you're already in the area, not as a separate refactor sprint

## Naming Matters

- If you need a comment to explain what a variable or function does, rename it instead
- Good names make comments unnecessary

## Minimize Dependencies

- Every dependency is a liability
- Prefer stdlib over third-party when the difference is small

## DRY — But Not Prematurely

- Eliminate actual duplication, not coincidental similarity
- Premature DRY leads to wrong abstractions — three similar lines are fine if the alternative is a forced generalization

## Make It Work, Make It Right, Make It Fast

- In that order — don't optimize before it's correct and clean
- Don't optimize until there's a measured performance problem
- Readability and correctness always come before speed

## Performance — Measure First, Optimize Never

- You can't predict where a program will spend its time — bottlenecks happen in surprising places
- Don't create speed hacks until you've measured where the bottlenecks actually are
- Even after measuring, don't tune unless one part of the program overwhelms the rest

## Don't Get Fancy

- Fancy algorithms are slow when n is small — and n is usually small
- Simple algorithms scale better than complex ones
- Fancy algorithms are buggier than simple ones — use simple algorithms for simple data
- Simpler is always easier to debug

## Data Dominates

- If you choose the right data structure, the algorithms almost write themselves
- Surround good data structures with dumb code — not the other way around
