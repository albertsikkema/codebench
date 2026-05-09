# Performance Optimization

## Principle

Don't optimize until you've measured. Bottlenecks happen in surprising places, simple algorithms beat fancy ones for real-world data sizes, and the right data structure matters more than clever code. Measure first, optimize only what dominates, and keep it simple.

## Why

- **You can't predict where time is spent**: Intuition about performance is unreliable. The bottleneck is almost never where you think it is. Speed hacks applied to the wrong code are wasted effort and added complexity.
- **Premature optimization obscures intent**: Code written for speed rather than clarity is harder to read, harder to debug, and harder to change. If the optimization wasn't necessary, you've made the codebase worse for nothing.
- **Fancy algorithms have hidden costs**: They carry complexity overhead (bugs, maintenance, cognitive load) that only pays off at scale. Most programs never reach that scale.

## Core Rules

### 1. Measure Before Tuning

Never optimize based on guesswork. Profile first, identify the actual bottleneck, then optimize only that part.

- Use profiling tools appropriate to your language (cProfile, pprof, Chrome DevTools, perf)
- If no single part of the program overwhelms the rest, there is nothing to optimize
- A 10x improvement to code that runs 1% of the time is invisible

### 2. Don't Get Fancy

Fancy algorithms are slow when n is small -- and n is usually small.

- O(n^2) with a small constant often beats O(n log n) with a large constant for n < 1000
- Simple algorithms are easier to implement correctly and easier to debug
- Only reach for complex algorithms when simple ones measurably fail
- If you can't explain the algorithm in one paragraph, it's probably overkill for the problem

### 3. Data Dominates

This is the most important rule. If you choose the right data structure and organize things well, the code around it will be simple. Dumb code that surrounds the right data structure will work better than clever code built around the wrong one.

- Start with the data: what are you storing, how are you accessing it, what are the invariants?
- Surround good data structures with dumb, straightforward code -- not the other way around
- A hash map with linear scans often beats a tree with clever traversal
- When performance matters, the data layout (cache locality, allocation patterns) usually matters more than the algorithm
- If your code is getting complicated, step back and ask whether the data structure is wrong
- Most "algorithm problems" are actually data structure problems in disguise
- Write the data structure first, then write the simplest possible code that operates on it

### 4. Optimize the Right Thing

When optimization is warranted, focus on the highest-impact change.

- Reduce I/O before reducing CPU -- network and disk are orders of magnitude slower
- Batch operations before parallelizing them
- Cache results before recomputing them faster
- Change the algorithm before micro-optimizing the implementation

## Anti-Patterns

| Anti-Pattern | Why It's Wrong | Do This Instead |
|---|---|---|
| Optimizing without profiling | You're probably optimizing the wrong thing | Profile, find the bottleneck, then optimize |
| Using a complex data structure "just in case" | Adds complexity for no measured benefit | Start simple, switch when measurements demand it |
| Micro-optimizing hot loops before fixing I/O | CPU work is rarely the bottleneck | Check I/O, allocation, and serialization first |
| Caching everything | Cache invalidation bugs are subtle and costly | Cache only what's measured to be slow and accessed often |
| Premature parallelism | Concurrency bugs are hard to find and fix | Sequential first, parallelize only when measurements justify it |
| Writing clever code around a bad data structure | The complexity never ends -- you're fighting the shape of your data | Redesign the data structure, then the code simplifies itself |

## References

- Rob Pike's Rules of Programming
- "Make It Work, Make It Right, Make It Fast" -- Kent Beck
