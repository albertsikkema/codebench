---
name: visual-verify
description: "Visual verification of web UI using Playwright MCP. Navigate, screenshot, inspect accessibility tree, and verify rendered output at multiple viewports."
---

# Visual Verification (Playwright MCP)

Verify rendered UI output using the Playwright MCP server. This skill closes the feedback loop after generating UI with front-end-design, ui-component-creator, or /build: render, screenshot, evaluate.

## 1. When to Use

- After generating UI with front-end-design or ui-component-creator
- After /build completes a frontend task
- Manual QA checks on rendered pages
- Verifying responsive behavior across viewports
- Checking accessibility tree structure

**Not for:** Unit testing, API testing, or non-browser verification.

## 2. Prerequisites

- **Playwright MCP server** must be configured (`@playwright/mcp` in `.mcp.json`)
- **Dev server** must be running and accessible at a known URL
- Install with: `./install-helper.sh --with-playwright`

## 3. Playwright MCP Tools Reference

| Tool | Purpose |
|------|---------|
| `mcp__playwright__browser_navigate` | Open URL in browser |
| `mcp__playwright__browser_snapshot` | Get accessibility tree (structure check) |
| `mcp__playwright__browser_take_screenshot` | Capture visual screenshot |
| `mcp__playwright__browser_resize` | Change viewport dimensions |
| `mcp__playwright__browser_click` | Click an element |
| `mcp__playwright__browser_type` | Type into an input |
| `mcp__playwright__browser_verify_text_visible` | Assert text is visible |
| `mcp__playwright__browser_verify_element_visible` | Assert element is visible |
| `mcp__playwright__browser_console_messages` | Check for JS errors |
| `mcp__playwright__browser_network_requests` | Verify API calls and status codes |
| `mcp__playwright__browser_evaluate` | Run JS for computed style checks |

## 4. Verification Workflow

### Step 1: Navigate

Open the target URL in the browser.

```
browser_navigate → target URL (e.g., http://localhost:3000/page)
```

### Step 2: Snapshot (Accessibility Tree)

Capture the accessibility tree to verify semantic structure — headings, landmarks, labels, roles.

```
browser_snapshot → review tree for correct heading hierarchy, landmark regions, ARIA labels
```

Check for:
- Proper heading levels (h1 > h2 > h3)
- Landmark regions (nav, main, footer)
- Form labels and ARIA attributes
- Interactive elements have accessible names

### Step 3: Screenshot (Visual Capture)

Take a screenshot to verify the visual output matches expectations.

```
browser_take_screenshot → review layout, spacing, typography, colors
```

Check for:
- Layout matches design intent
- No visual overflow or clipping
- Text is readable and properly sized
- Colors and contrast look correct
- No missing images or broken assets

### Step 4: Responsive Viewports

Resize and screenshot at three standard viewports:

| Viewport | Width | Height | Represents |
|----------|-------|--------|------------|
| Desktop | 1920 | 1080 | Full HD desktop |
| Tablet | 768 | 1024 | iPad portrait |
| Mobile | 375 | 667 | iPhone SE |

For each viewport:
```
browser_resize → set dimensions
browser_take_screenshot → capture
```

Check for:
- Layout adapts correctly (columns collapse, nav changes)
- No horizontal scrolling on mobile
- Touch targets are adequately sized
- Content remains readable
- No elements overflow the viewport

### Step 5: Interact and Assert

Test key interactions by clicking buttons, typing in forms, and verifying expected outcomes.

```
browser_click → target element
browser_type → input field
browser_verify_text_visible → expected text appears
browser_verify_element_visible → expected element appears
```

Test at minimum:
- Primary CTA / main action works
- Navigation links work
- Form submission flows (if applicable)
- Modal/dialog open and close

## 5. Console and Network Checks

### Console Messages

```
browser_console_messages → check for JS errors
```

- **Errors**: Any JS error is a blocker — investigate and fix
- **Warnings**: Review for deprecations or misuse (React key warnings, etc.)
- **Info/debug**: Generally safe to ignore

### Network Requests

```
browser_network_requests → verify API calls
```

Check for:
- No failed fetches (network errors)
- No unexpected 4xx responses (404 missing assets, 401 auth failures)
- No 5xx server errors
- All expected API calls completed successfully
- Static assets (CSS, JS, images) loaded without errors

## 6. Output Format

After verification, produce a markdown report:

```markdown
## Visual Verification Report

**URL**: http://localhost:3000/page
**Date**: YYYY-MM-DD

### Verdict: PASS / FAIL / PARTIAL

### Viewport Results

| Viewport | Status | Notes |
|----------|--------|-------|
| Desktop (1280x720) | PASS | Layout correct |
| Tablet (768x1024) | PASS | Nav collapses properly |
| Mobile (375x667) | FAIL | CTA button overflows viewport |

### Accessibility Tree
- Heading hierarchy: OK
- Landmarks: OK
- Form labels: 1 missing label on search input

### Console
- Errors: None
- Warnings: 1 React key warning in list component

### Network
- Failed requests: None
- All API calls returned 2xx

### Issues Found
1. **[Mobile]** CTA button width exceeds viewport at 375px — needs `max-w-full`
2. **[A11y]** Search input missing `aria-label` or associated `<label>`
3. **[Warning]** React key warning in ProductList component
```

## 7. Pre-Flight Checklist

- [ ] Dev server is running and accessible
- [ ] Playwright MCP server is configured
- [ ] Target URL is known and reachable
- [ ] Page has finished loading (no spinners/skeletons)

## 8. Post-Verification Checklist

- [ ] Tested at all three viewports (desktop, tablet, mobile)
- [ ] Accessibility tree reviewed for semantic correctness
- [ ] Console checked for JS errors
- [ ] Network requests verified (no failures, no unexpected 4xx/5xx)
- [ ] Key interactions tested (clicks, form input, navigation)
- [ ] Report generated with verdict and issues
- [ ] Issues filed or fixed before marking complete

## 9. Integration

| Skill | Provides | This skill adds |
|-------|----------|----------------|
| **front-end-design** | Creative direction, visual design | Visual verification that output matches design intent |
| **ui-component-creator** | Component structure, variants | Verification that components render correctly |
| **mobile-friendly-design** | Responsive patterns, breakpoints | Multi-viewport screenshot testing |
| **accessibility** | WCAG patterns, ARIA, keyboard nav | Accessibility tree snapshot verification |
