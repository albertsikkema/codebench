# Accessibility

## Principle

Build interfaces that work for everyone — including people who navigate with keyboards, use screen readers, have low vision, or experience motion sensitivity. Accessibility is not a feature to add later; it's a quality of the HTML, CSS, and interaction design from the start. Target WCAG 2.2 Level AA.

## Why

- **It's the law in many jurisdictions**: The European Accessibility Act (2025), ADA (US), EN 301 549 (EU) all require digital accessibility. Non-compliance creates legal risk.
- **Accessible design is better design**: Captions help in noisy environments. Keyboard navigation helps power users. High contrast helps in sunlight. Good semantics help SEO. Accessibility improvements benefit everyone.
- **Retrofitting is expensive**: Fixing accessibility after the UI is built means rewriting components, restructuring HTML, and re-testing everything. Building it in from the start costs almost nothing.

## Core Rules

### 1. Use Semantic HTML

Use the right HTML element for each purpose. Semantic elements carry meaning that assistive technologies rely on.

| Purpose | Correct element | Wrong element |
|---------|---------------|---------------|
| Navigation | `<nav>` | `<div class="nav">` |
| Primary content | `<main>` | `<div id="content">` |
| Button/action | `<button>` | `<div onClick={...}>` |
| Link to another page | `<a href="...">` | `<span onClick={...}>` |
| Page header | `<header>` | `<div class="header">` |
| Content section | `<section>`, `<article>` | `<div>` |

**Rule**: Use `<div>` and `<span>` only for layout/styling wrappers with no semantic meaning. If the element does something or means something, there's a semantic element for it.

### 2. Make Everything Keyboard Accessible

All interactive elements must be reachable via Tab and operable via Enter/Space.

**Requirements**:
- Tab order follows visual layout (don't rearrange with `tabindex` values > 0)
- No keyboard traps (user can always Tab out of any component)
- Focus indicator is visible and meets WCAG 2.2 Focus Appearance criteria
- Skip-to-content link on every page (first focusable element)
- Modal dialogs trap focus within the modal and return focus when closed

**Testing**: Unplug your mouse. Navigate the entire interface with Tab, Shift+Tab, Enter, Space, Escape, and Arrow keys. If you can't complete a task, it's inaccessible.

### 3. Provide Text Alternatives

Every non-text element needs a text equivalent:

| Element | Text alternative |
|---------|-----------------|
| Informative image | `alt="Description of what the image shows"` |
| Decorative image | `alt=""` (empty alt, not missing alt) |
| Icon button | `aria-label="Close"` or visually hidden text |
| SVG icon | `role="img"` with `aria-label` or `<title>` |
| Chart/graph | Data table or text summary |
| Video | Captions and transcript |

**Rule**: If you remove all images and CSS, can a user still understand the content and complete all tasks? If not, text alternatives are missing.

### 4. Meet Color Contrast Requirements

| Element type | Minimum contrast ratio |
|-------------|----------------------|
| Normal text (< 18pt / 14pt bold) | 4.5:1 |
| Large text (≥ 18pt / 14pt bold) | 3:1 |
| UI components (buttons, inputs, icons) | 3:1 against adjacent colors |

**Never rely on color alone** to convey information. Pair color with:
- Icons (error icon + red color)
- Patterns (striped + colored bars in charts)
- Text labels ("Required" + red asterisk)

**Both themes must pass**: If your app has light and dark modes, both must meet contrast requirements.

### 5. Label Every Form Input

Every `<input>`, `<select>`, and `<textarea>` must have a programmatically associated `<label>`.

```html
<!-- GOOD: label with for/id -->
<label for="email">Email address</label>
<input id="email" type="email" required aria-required="true">

<!-- GOOD: wrapping label -->
<label>
  Email address
  <input type="email" required>
</label>

<!-- BAD: no association -->
<span>Email address</span>
<input type="email">
```

**Additional requirements**:
- Required fields: indicate with more than color (asterisk + `aria-required="true"`)
- Validation errors: link to field via `aria-describedby`, display as text
- Related fields: group with `<fieldset>` and `<legend>`

### 6. Respect Motion Preferences

```css
/* Reduce or remove animations when user prefers reduced motion */
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

**Rules**:
- No content auto-plays without a pause/stop mechanism
- Nothing flashes more than 3 times per second
- Carousels and marquees have visible controls
- Parallax scrolling is disabled in reduced-motion mode

### 7. Handle Dynamic Content

Single-page applications and dynamic updates need special attention:

**Route changes in SPAs**:
- Move focus to the main content area or new page heading
- Announce the new page title to screen readers
- Update `document.title`

**Dynamic updates** (toasts, form results, loading states):
- Use `aria-live="polite"` for non-urgent updates
- Use `aria-live="assertive"` for errors and critical alerts
- Loading indicators need `aria-busy="true"` and a text label

```html
<!-- Toast notification region -->
<div aria-live="polite" aria-atomic="true" class="sr-only" id="notifications">
  <!-- Dynamically inserted messages are announced to screen readers -->
</div>
```

### 8. Support Text Resizing

Layout must reflow without horizontal scrolling at 200% browser zoom (up to 1280px viewport width).

**Rules**:
- Use relative units (`rem`, `em`, `%`) for font sizes, not `px`
- Don't disable user scaling (`maximum-scale=1` in viewport meta)
- Test at browser zoom 200% and OS font scaling 200%
- Content must remain readable and functional at increased sizes

### 9. Provide Navigation Aids

- Every page has a unique, descriptive `<title>`
- Consistent navigation placement across pages
- At least two ways to reach any page (nav menu + search, or nav + sitemap)
- Use `<nav>` with `aria-label` when multiple navigation regions exist
- Active page/section indicated visually and programmatically

### 10. Run Automated Checks in CI

Automated tools catch ~30-40% of accessibility issues. They are necessary but not sufficient.

**Tools**:
- **axe-core**: Run in CI on rendered pages. Fail on AA violations.
- **Lighthouse accessibility audit**: Part of Lighthouse CI.
- **eslint-plugin-jsx-a11y**: Catch issues at build time in React projects.

**What automated tools cannot check**:
- Whether focus order is logical
- Whether text alternatives are meaningful (not just "image")
- Whether keyboard interaction patterns are intuitive
- Whether screen reader announcements make sense in context

**Complement with manual testing**: Navigate with keyboard only. Test with a screen reader (VoiceOver on macOS, NVDA on Windows). Verify at 200% zoom.

## Implementation Patterns

### Skip-to-Content Link

```html
<body>
  <a href="#main-content" class="skip-link">Skip to main content</a>
  <nav>...</nav>
  <main id="main-content">...</main>
</body>

<style>
.skip-link {
  position: absolute;
  top: -40px;
  left: 0;
  padding: 8px 16px;
  background: #000;
  color: #fff;
  z-index: 1000;
}
.skip-link:focus {
  top: 0; /* visible only when focused */
}
</style>
```

### Visually Hidden Text

```css
/* Accessible to screen readers but not visible */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
```

Use for icon buttons, decorative headings, or context that screen readers need but visual users don't.

### Accessible Modal Dialog

```html
<dialog id="confirm-dialog" aria-labelledby="dialog-title" aria-modal="true">
  <h2 id="dialog-title">Confirm deletion</h2>
  <p>Are you sure you want to delete this item?</p>
  <button autofocus>Cancel</button>
  <button>Delete</button>
</dialog>
```

The native `<dialog>` element with `showModal()` handles focus trapping, Escape to close, and background inertia automatically. Prefer it over custom modal implementations.

### Touch Targets

```css
/* Minimum 44x44px touch targets */
button, a, input[type="checkbox"] + label, input[type="radio"] + label {
  min-height: 44px;
  min-width: 44px;
}

/* Minimum 8px spacing between adjacent targets */
.button-group > * + * {
  margin-left: 8px;
}
```

## When to Bend the Rules

- **Admin-only internal tools**: WCAG AA is still recommended, but the legal risk is lower. At minimum: keyboard navigation, semantic HTML, and color contrast.
- **Data visualizations**: Complex charts may not be fully accessible. Provide a data table alternative and a text summary. The chart itself can be `role="img"` with an `aria-label`.
- **Third-party embedded content** (maps, video players): You can't control their accessibility. Provide alternatives (text address alongside map, transcript alongside video).
- **Rapid prototypes**: Use semantic HTML from the start (costs nothing extra). Skip ARIA live regions and advanced keyboard patterns until the UI stabilizes.
