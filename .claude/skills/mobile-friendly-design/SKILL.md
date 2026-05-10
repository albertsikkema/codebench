---
name: mobile-friendly-design
description: "Use when building responsive web interfaces that must work across phone, tablet, and desktop. Covers mobile navigation, touch-friendly interactions, responsive layouts, and mobile form patterns. Complements ui-component-creator (structure) and front-end-design (creative direction). Not for native mobile apps — responsive web only."
---

# Mobile-Friendly Design

Responsive web patterns for phone, tablet, and desktop. This skill adds the **responsive layer** on top of component structure (ui-component-creator) and creative direction (front-end-design).

## 1. When to Use / When NOT to Use

**Use this skill when:**
- Building responsive layouts that must work on phone + tablet + desktop
- Implementing mobile navigation (hamburger, drawer, bottom tabs)
- Adding touch-friendly interactions to web components
- Building mobile-optimized forms

**Do NOT use for:**
- React Native, Flutter, or native mobile apps — this is responsive web only
- Defining breakpoint tokens or spacing scales — those live in the project's `design-system.md`
- Component structure (forwardRef, cn(), variants) — use `ui-component-creator`
- Creative/aesthetic direction — use `front-end-design`

**Works with** ui-component-creator and front-end-design, not instead of.

## 2. Do / Don't Quick Reference

```tsx
// ❌ Desktop-first class order
<div className="grid grid-cols-3 md:grid-cols-2 sm:grid-cols-1">

// ✅ Mobile-first class order
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
```

```tsx
// ❌ Fixed widths that break on small screens
<div className="w-[800px]">

// ✅ Fluid widths with max constraint
<div className="w-full max-w-3xl">
```

```tsx
// ❌ Hover-only information
<div className="opacity-0 hover:opacity-100">Delete</div>

// ✅ Always visible on touch, enhanced on hover
<div className="opacity-100 md:opacity-0 md:group-hover:opacity-100">Delete</div>
```

```tsx
// ❌ Tiny tap targets
<button className="p-1 text-xs">×</button>

// ✅ 44x44px minimum touch target
<button className="min-h-[44px] min-w-[44px] p-2">×</button>
```

```tsx
// ❌ Desktop grid forced on mobile
<div className="grid grid-cols-4 gap-8">

// ✅ Progressive grid
<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 lg:gap-8">
```

```tsx
// ❌ Disabled zoom (accessibility violation)
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1" />

// ✅ Accessible viewport
<meta name="viewport" content="width=device-width, initial-scale=1" />
```

```tsx
// ❌ Fixed bottom without safe area
<div className="fixed bottom-0">

// ✅ Fixed bottom with safe area
<div className="fixed bottom-0 pb-[env(safe-area-inset-bottom)]">
```

```tsx
// ❌ Desktop nav on mobile (horizontal overflow)
<nav className="flex gap-8">{allLinks}</nav>

// ✅ Responsive nav: mobile drawer, desktop horizontal
<nav className="hidden md:flex gap-6">{allLinks}</nav>
<MobileDrawer className="md:hidden" />
```

## 3. Breakpoint Usage Guide

Don't redefine breakpoint tokens here — those live in the project's `design-system.md`. This section covers **what typically changes** at each breakpoint.

| Breakpoint | What typically changes |
|------------|----------------------|
| **base** (0+) | Single column, stacked layout, full-width elements, hamburger nav |
| **sm** (640px) | Minor tweaks, 2-col where needed, slightly larger tap targets |
| **md** (768px) | Sidebar appears, 2-col grids, horizontal nav replaces hamburger |
| **lg** (1024px) | 3-col grids, expanded sidebar, more horizontal space |
| **xl** (1280px) | Max content width, luxury spacing, multi-panel layouts |

Always design mobile-first: start at base and add complexity upward.

## 4. Layout Patterns (Quick Reference)

**Responsive container:**
```tsx
<div className="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8">
  {children}
</div>
```

**Responsive grid (1→2→3 columns):**
```tsx
<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 lg:gap-6">
  {items.map(item => <Card key={item.id} {...item} />)}
</div>
```

**Full-width mobile cards:**
```tsx
<div className="-mx-4 sm:mx-0 sm:rounded-lg">
  <div className="px-4 py-3 sm:px-6">{content}</div>
</div>
```

See `references/layout-patterns.md` for sidebar+content, stacking order, responsive spacing, and typography patterns.

## 5. Interaction Patterns (Quick Reference)

**Navigation by app type:**

| App type | Phone | Tablet+ |
|----------|-------|---------|
| Content/marketing | Hamburger → drawer | Horizontal nav |
| Dashboard/SaaS | Bottom tabs or hamburger | Collapsible sidebar |
| E-commerce | Bottom tabs + hamburger | Mega menu |

**Touch target sizing:**
```tsx
// Minimum 44x44px for all interactive elements
<button className="min-h-[44px] min-w-[44px] flex items-center justify-center p-2">
  <Icon className="h-5 w-5" />
</button>
```

**Mobile form input types:**

| Data | `type` | `inputMode` | Why |
|------|--------|-------------|-----|
| Email | `email` | — | Shows @ on keyboard |
| Phone | `tel` | — | Numeric pad |
| Amount | `text` | `decimal` | Numeric with decimal |
| Search | `search` | — | Shows search action key |
| URL | `url` | — | Shows .com / slash |

See `references/interaction-patterns.md` for drawer components, bottom tabs, hamburger menus, tap feedback, and autocomplete patterns.

## 6. Performance Essentials

**Viewport meta (required):**
```html
<meta name="viewport" content="width=device-width, initial-scale=1" />
```
Never set `maximum-scale=1` — it disables pinch-to-zoom and is an accessibility violation.

**Responsive images (Next.js):**
```tsx
<Image
  src="/hero.jpg"
  alt="Hero"
  sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
  fill
  className="object-cover"
/>
```

**Reduced motion:**
```tsx
<div className="animate-fade-in motion-reduce:animate-none">
```

**Safe areas (notched devices):**
```tsx
// Requires Tailwind config extension for env() support
<div className="pb-[env(safe-area-inset-bottom)]">
```

Add `viewport-fit=cover` to the viewport meta to enable safe area insets:
```html
<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover" />
```

## 7. Mobile-Readiness Checklist

- [ ] Base styles are mobile layout (single column, stacked)
- [ ] No fixed widths that break on small screens
- [ ] Responsive grid adapts per breakpoint
- [ ] Mobile nav is reachable (drawer, hamburger, or bottom tabs)
- [ ] All touch targets ≥ 44x44px
- [ ] No essential info behind hover-only interactions
- [ ] Form inputs use correct type (email, tel, url)
- [ ] Viewport meta is correct (no maximum-scale)
- [ ] Fixed bottom elements respect safe-area-inset-bottom
- [ ] Animations respect prefers-reduced-motion
- [ ] Tested at 375px (phone) and 768px (tablet)
- [ ] Touch works without hover

## 8. Integration

| Skill | Provides | This skill adds |
|-------|----------|----------------|
| Project `design-system.md` | Breakpoint tokens, spacing scale | How and when to use those breakpoints |
| **ui-component-creator** | forwardRef, cn(), variants, a11y | Responsive variants, touch-sizing |
| **front-end-design** | Creative direction, aesthetics | How that vision translates responsively |
| **accessibility** | WCAG compliance, ARIA, keyboard nav | Touch target sizing, form labeling, zoom policy |
