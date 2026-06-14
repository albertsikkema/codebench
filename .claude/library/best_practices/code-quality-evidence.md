# Code Quality Analysis Tools -- Evidence Guide

What the research supports, how to weight findings, and which tools to use.

## Evidence Tiers

Each analysis method below is classified by strength of empirical evidence. Use this to prioritize findings: a `[strong]` finding outweighs a `[moderate]` one; a `[practitioner]` finding is worth flagging but should not block a merge on its own.

| Tier | Label | Meaning |
|------|-------|---------|
| 1 | **[strong]** | Replicated peer-reviewed studies with controlled confounds |
| 2 | **[moderate]** | Peer-reviewed but with known confounds (e.g., LOC correlation) or limited replication |
| 3 | **[practitioner]** | Widely accepted in practice, not empirically isolated in controlled studies |

## The LOC Confound

Most code metrics correlate with lines of code. El Emam et al. (2001) showed that out of 24 OO metrics, only 4 retained any relationship to faults after controlling for class size. Jay et al. (2009) showed cyclomatic complexity has "absolutely no explanatory power of its own" beyond LOC. Any metric that claims to be useful must demonstrate predictive power after controlling for size.

This does not make size-correlated metrics useless for maintainability -- a 500-line function IS harder to maintain -- but it means the metric is not finding something LOC alone would miss. Keep this in mind when interpreting results.

---

## Methods and MCP Tools

### Hotspot Analysis (Change Frequency x Complexity)

**Evidence: [strong]**

Files that change often AND are complex are where problems concentrate. Nagappan and Ball (2005) showed relative churn discriminated fault-prone binaries at 89% accuracy. Tornhill and Borg (2022) found low-quality hotspot code contains 15x more defects and takes 124% longer to work on.

**MCP tool**: `find_hotspots(since?, min_changes?, limit?)` -- combines git change frequency with code complexity.

**Weight**: High. Hotspot findings should be treated as high-priority signals. If a changed file is a hotspot, scrutinize it harder.

### Code Ownership and Knowledge Distribution

**Evidence: [strong]**

Bird et al. (2011) found that components with many low-expertise contributors had significantly more defects. Removing ownership features from their prediction model dramatically decreased performance.

**MCP tool**: `find_ownership_risks()` -- identifies distributed ownership (Bird et al.) and bus-factor=1 files via git history.

**Weight**: High for maintainability and risk assessment. Bus-factor=1 files touched in a PR deserve extra review attention.

### Cognitive Complexity

**Evidence: [moderate]**

Munoz Baron et al. (2020) confirmed correlation with perceived understandability. Lenarduzzi et al. (2023) found it slightly outperforms cyclomatic complexity for readability. NOT validated as a defect predictor. May be a LOC proxy (Shepperd, 1988).

**MCP tool**: `find_cognitive_complexity(threshold?, language?, file?, limit?)`

**Weight**: Medium. Useful as maintainability signal. A function scoring 47 is harder to understand than one scoring 8. Not a bug predictor.

### Deep Nesting

**Evidence: [moderate]**

Hatton (1997) identified nesting depth as important for defect variability. Confounded with function length (El Emam et al., 2001). No evidence that reducing nesting alone improves outcomes.

**MCP tool**: `find_deep_nesting(threshold?, language?, file?, limit?)`

**Weight**: Medium. Flag at 3-4+ levels. Better as a threshold signal than continuous measurement. Supporting signal, not primary.

### Code Smells (Bloated Functions, Feature Envy)

**Evidence: [moderate] for god class/long method, [weak] for exotic smells**

Palomba et al. (2018) found positive correlation between smells and bugs. But El Emam's size confound applies. Sharma and Spinellis (2018) found evidence varies widely per smell type. Feature envy shows weaker, less consistent results.

**MCP tools**:
- `find_bloated_functions(threshold?, language?, file?, limit?)` -- [moderate] evidence
- `find_feature_envy(threshold?, language?, file?, limit?)` -- [weak] evidence
- `find_nested_loop_patterns(language?, file?, limit?)` -- [practitioner]

**Weight**: Medium for bloated functions (size is a real maintainability problem). Low for feature envy and nested loops -- flag but don't block.

### Temporal Coupling (Co-Change Analysis)

**Evidence: [moderate]**

D'Ambros et al. (2009) found change coupling correlates with defects. Canfora et al. (2014) found 64-93% of defects in Granger-positive classes. Noisy: bulk refactoring and API changes produce false positives.

**MCP tool**: `find_temporal_coupling()` -- detects hidden co-change dependencies between files with no import relationship.

**Weight**: Medium. Unique signal no other method provides. Use to identify hidden dependencies. Filter aggressively, present as "hidden dependency finder."

### Duplicate Code

**Evidence: [moderate], conflicting**

Juergens et al. (2009): 52% of clones inconsistently changed, 15% of those caused faults. But Bettenburg et al. (2012): only 1-3% of inconsistencies cause defects. Rahman et al. (2012): clones may actually be less defect-prone.

**MCP tool**: `find_duplicates(threshold?, language?, min_lines?, limit?)`

**Weight**: Medium for maintainability (fix one copy, forget the others). Not a defect predictor. Useful for reducing future maintenance surface.

### Coupling Metrics

**Evidence: [moderate]**

Al Dallal (2013) found Martin's coupling metrics lack theoretical/empirical evaluation. Chidamber-Kemerer CBO (1994) has better validation. May be LOC proxy.

**MCP tool**: `analyze_coupling(file_or_module?)` -- Ca (afferent), Ce (efferent), instability metrics.

**Weight**: Medium for maintainability. A module with 30 incoming dependencies is risky to change. Simple coupling counts suffice; fancy frameworks add little.

### Circular Dependencies

**Evidence: [practitioner]**

Limited empirical evidence linking cycles to defects. But detection is deterministic with zero false positives.

**MCP tool**: `find_circular_deps()`

**Weight**: Low for defect prediction, high confidence for architectural health. Zero false positives makes it safe to always flag.

### Error Handling Analysis

**Evidence: [practitioner]**

No large-scale empirical study quantifying defect rates from specific patterns. High face validity: ignored error returns are bugs waiting to happen. Low false positive rate.

**MCP tool**: `find_unhandled_errors(language?, file?, limit?)`

**Weight**: Medium-high. Both a bug finder (swallowed errors are real bugs) and maintainability signal. Low false positive rate makes it safe to flag.

### Testability Analysis

**Evidence: [moderate]**

Bruntink and van Deursen (2006): WMC (complexity) had strongest correlation with test effort. Size consistently the biggest single predictor. Hevery's Google testability guide identifies four flaw categories: constructor complexity, Law of Demeter violations, global state, class does too much. NDSS 2022 (Alkassar et al.): testability improvements led to discovery of 440 new vulnerabilities -- strongest causal evidence in the testability space.

**MCP tool**: `find_testability_issues()` -- detects constructor complexity, concrete dependency instantiation, global state access, Law of Demeter violations, hard-to-isolate functions.

**Weight**: Medium. Frame as "code that will be hard to test," not "code that has bugs." Overlap with complexity/coupling checks. Focus on high-confidence patterns where false positive rate is naturally low.

---

## Methods NOT Recommended

### Composite Health Scores
**Evidence: [negative]**. Combining metrics on different scales into one number violates measurement theory (Fenton and Pfleeger, 1997). A "7.2 health score" hides which dimensions are suffering. Show individual signals instead.

### Function Length (standalone)
**Evidence: [negative]**. Does not predict defects beyond what LOC already tells you. Covered by bloated functions and complexity checks.

---

## How to Use This in Reviews

1. **Prioritize by evidence tier**: [strong] findings first, then [moderate], then [practitioner]
2. **Hotspots and ownership risks are the highest-value signals** -- if a changed file is a hotspot or has ownership risk, flag it prominently
3. **Complexity, nesting, coupling are supporting signals** -- they tell you WHERE to look harder, not that a bug exists
4. **Error handling is actionable** -- swallowed errors and unchecked returns are concrete bugs
5. **Testability issues inform test review** -- code flagged as hard to test should have extra test scrutiny
6. **Don't block on [practitioner]-only findings** unless the pattern is clearly a bug (e.g., bare except swallowing all errors)
7. **Remember Goodhart's Law**: when a metric becomes a target, it stops being useful. Don't optimize for the metric; use it as a signal

## Key Sources

- Nagappan & Ball (2005) -- churn predicts defects at 89% accuracy
- Bird et al. (2011) -- code ownership predicts defects
- El Emam et al. (2001) -- most metrics are LOC proxies
- Jay et al. (2009) -- CC has no explanatory power beyond LOC
- Tornhill & Borg (2022) -- 15x defects in low-quality hotspot code
- Bruntink & van Deursen (2006) -- complexity strongest testability predictor
- Palomba et al. (2018) -- smells correlate with bugs (with size caveats)
- Bessey et al. (2010) -- false positives kill adoption
- D'Ambros et al. (2009) -- change coupling correlates with defects
- NDSS 2022, Alkassar et al. -- testability improvements find real vulnerabilities
