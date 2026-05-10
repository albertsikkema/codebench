---
name: accessibility
description: "WCAG 2.2 compliance patterns for web interfaces. Use when building or reviewing UI for accessibility — semantic HTML, ARIA, keyboard navigation, color contrast, motion safety, screen reader support, and automated testing. Complements ui-component-creator (structure), front-end-design (aesthetics), and mobile-friendly-design (responsive)."
---

# Accessibility (WCAG 2.2)

Patterns and rules for building accessible web interfaces. This skill provides the **accessibility layer** on top of component structure (ui-component-creator), creative direction (front-end-design), and responsive design (mobile-friendly-design).

## 1. When to Use

- Building new UI components or pages
- Reviewing existing code for accessibility compliance
- Adding accessibility to components that lack it
- Preparing for accessibility audits

**Not for:** Native mobile apps (React Native, Flutter) — web only.

## 2. WCAG Levels Quick Reference

| Level | Target | When to use |
|-------|--------|-------------|
| **A** | Minimum baseline | Always — non-negotiable |
| **AA** | Standard target | **Default for all projects** — covers contrast, keyboard, labels |
| **AAA** | Enhanced | When required by contract, government, or user need |

Default to **AA** unless the project specifies otherwise.

## 3. Semantic HTML

Use the correct element for the job — ARIA should supplement, not replace, native semantics.

```html
<!-- ❌ Div soup -->
<div class="header">
  <div class="nav">
    <div onclick="navigate()">Home</div>
  </div>
</div>

<!-- ✅ Semantic elements -->
<header>
  <nav aria-label="Main">
    <a href="/">Home</a>
  </nav>
</header>
```

**Landmark regions**: Use `<header>`, `<nav>`, `<main>`, `<aside>`, `<footer>`. Add `aria-label` when multiple landmarks of the same type exist.

**Heading hierarchy**: One `<h1>` per page. Never skip levels (h1 → h3). Headings describe section content.

## 4. ARIA Patterns

**First rule of ARIA**: Don't use ARIA if a native HTML element does the job.

| Pattern | Key attributes | Notes |
|---------|---------------|-------|
| Dialog/Modal | `role="dialog"`, `aria-modal="true"`, `aria-labelledby` | Trap focus, return focus on close |
| Tabs | `role="tablist/tab/tabpanel"`, `aria-selected`, `aria-controls` | Arrow keys navigate tabs |
| Accordion | `<button aria-expanded>`, `aria-controls` | Native `<details>` often sufficient |
| Combobox | `role="combobox"`, `aria-expanded`, `aria-activedescendant` | Complex — see references |
| Alert | `role="alert"` or `aria-live="assertive"` | Announced immediately |
| Status | `role="status"` or `aria-live="polite"` | Announced at next pause |

**Live regions**: Use `aria-live="polite"` for non-urgent updates (search results count, save confirmation). Use `aria-live="assertive"` for urgent messages (errors, session expiry).

See `references/aria-patterns.md` for full code examples.

## 5. Keyboard Navigation

Every interactive element must be operable with keyboard alone.

**Focus management rules:**
- Interactive elements are focusable by default (`<a>`, `<button>`, `<input>`, `<select>`, `<textarea>`)
- Use `tabindex="0"` to add custom elements to tab order
- Use `tabindex="-1"` for programmatic focus (not in tab order)
- **Never** use `tabindex` > 0

**Focus trapping (modals/dialogs):**
```tsx
// Trap focus inside modal — Tab/Shift+Tab cycle within
// On open: focus first focusable element
// On close: return focus to trigger element
```

**Roving tabindex (tab panels, toolbars, menus):**
```tsx
// Only one item in the group has tabindex="0"
// Arrow keys move tabindex="0" between items
// Other items have tabindex="-1"
```

**Skip link** (first element in the page):
```html
<a href="#main-content" class="sr-only focus:not-sr-only focus:absolute focus:z-50 focus:p-4">
  Skip to main content
</a>
```

## 6. Color & Contrast

| Element | Minimum ratio (AA) | Enhanced (AAA) |
|---------|-------------------|----------------|
| Normal text (< 18px) | **4.5:1** | 7:1 |
| Large text (≥ 18px bold or ≥ 24px) | **3:1** | 4.5:1 |
| UI components & graphical objects | **3:1** | — |

**Rules:**
- Never convey information by color alone — add icons, patterns, or text
- Support `prefers-contrast: more` for users who need higher contrast
- Test with grayscale filter to verify non-color indicators work

```css
@media (prefers-contrast: more) {
  :root {
    --border: oklch(0.4 0 0); /* Stronger borders */
  }
}
```

## 7. Motion & Animation

```css
/* Always gate non-essential animations */
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}
```

**Safe defaults:**
- Opacity fades and color transitions are generally safe
- Gate: parallax, auto-playing carousels, decorative motion, bouncing/spinning
- Never auto-play video with motion; provide pause controls

**Tailwind shorthand:**
```tsx
<div className="animate-fade-in motion-reduce:animate-none">
```

## 8. Screen Readers

**Visually hidden text** (announced but not visible):
```tsx
<span className="sr-only">Close dialog</span>
```

**Alt text rules:**
- Informative images: describe the content (`alt="Bar chart showing Q4 revenue growth"`)
- Decorative images: empty alt (`alt=""`) or use CSS background
- Icons with text labels: `aria-hidden="true"` on the icon
- Icon-only buttons: `aria-label` on the button

**Form labeling:**
```tsx
// ❌ Placeholder as label
<input placeholder="Email" />

// ✅ Visible label
<label htmlFor="email">Email</label>
<input id="email" type="email" />

// ✅ sr-only label (when design requires no visible label)
<label htmlFor="search" className="sr-only">Search</label>
<input id="search" type="search" placeholder="Search..." />
```

## 9. Forms

- Every input needs a `<label>` (visible or sr-only)
- Mark required fields with `aria-required="true"` and visible indicator
- Link error messages to inputs with `aria-describedby`
- Announce validation errors with `aria-live="polite"` or `role="alert"`

```tsx
<div>
  <label htmlFor="email">
    Email <span aria-hidden="true">*</span>
  </label>
  <input
    id="email"
    type="email"
    aria-required="true"
    aria-invalid={hasError}
    aria-describedby={hasError ? "email-error" : undefined}
  />
  {hasError && (
    <p id="email-error" role="alert" className="text-destructive text-sm">
      Please enter a valid email address.
    </p>
  )}
</div>
```

**Form groups**: Use `<fieldset>` + `<legend>` for related inputs (radio groups, address fields).

## 10. Testing Checklist

- [ ] **Keyboard-only walkthrough**: Tab through entire page, all controls reachable and operable
- [ ] **Focus visible**: Focus indicator visible on every interactive element
- [ ] **Screen reader test**: Navigate with VoiceOver (Mac) or NVDA (Windows) — headings, landmarks, forms, dynamic content announced correctly
- [ ] **axe-core / Lighthouse**: Run automated scan, zero critical/serious violations
- [ ] **Contrast checker**: All text and UI components meet AA ratios
- [ ] **Zoom test**: Content usable at 200% zoom, no horizontal scrolling
- [ ] **Reduced motion**: Enable `prefers-reduced-motion`, verify animations stop
- [ ] **No color-only info**: Information conveyed without relying solely on color
- [ ] **Form errors**: Error messages programmatically linked and announced

## 11. Do / Don't Quick Reference

| Do | Don't |
|----|-------|
| Use `<button>` for actions | Use `<div onClick>` for clickable elements |
| Use `<a href>` for navigation | Use `<span onClick>` for links |
| Provide visible focus styles | Use `outline: none` without replacement |
| Use `aria-label` on icon-only buttons | Rely on `title` attribute for accessibility |
| Gate animations with `prefers-reduced-motion` | Auto-play motion without pause control |
| Use `aria-live` for dynamic content | Update DOM silently without announcement |
| Label all form inputs | Use placeholder as the only label |
| Use semantic headings (h1-h6) | Use styled `<div>` or `<span>` for headings |
| Test with keyboard and screen reader | Rely solely on automated tools |
| Provide text alternatives for images | Use images of text instead of real text |

## 12. Integration

| Skill | Provides | This skill adds |
|-------|----------|----------------|
| **ui-component-creator** | forwardRef, cn(), variants, focus-visible | ARIA patterns, keyboard nav, screen reader support |
| **front-end-design** | Creative direction, color, typography | Contrast compliance, motion safety, non-color indicators |
| **mobile-friendly-design** | Touch targets, responsive layout | Touch target a11y (44px min), zoom policy, form input types |
| Project `design-system.md` | Color tokens, spacing scale | Contrast validation, focus token definitions |

## Record Completion (Audit Mode)

When this skill is used to **review or audit** existing code for accessibility (not when building new components), record the health check so the team session knows when it was last run:

1. Get the current HEAD commit hash: run `git rev-parse HEAD`
2. Get the current `health_checks` document via MCP: `get_document(type="health_checks")`
   - If it doesn't exist yet, start with an empty JSON object `{}`
3. Update the `accessibility` key with the current commit, timestamp, and a one-line summary of findings
4. Save via MCP: `update_document(type="health_checks", content=<updated JSON>)`
