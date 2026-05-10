# Interaction Patterns Reference

Concrete Tailwind CSS + React patterns for mobile navigation, touch interactions, and forms.

---

## Navigation

### Responsive Sidebar → Drawer

Full sidebar on desktop, slide-out drawer on mobile. Uses a dialog-based pattern compatible with Radix/Headless UI.

```tsx
import { useState } from "react";
import { cn } from "@/lib/utils";

interface SidebarProps {
  children: React.ReactNode;
  nav: React.ReactNode;
}

export function ResponsiveSidebar({ children, nav }: SidebarProps) {
  const [open, setOpen] = useState(false);

  return (
    <div className="flex min-h-screen">
      {/* Desktop sidebar */}
      <aside className="hidden md:flex md:w-64 md:flex-col md:border-r">
        <nav className="flex-1 overflow-y-auto p-4">{nav}</nav>
      </aside>

      {/* Mobile drawer */}
      {open && (
        <div className="fixed inset-0 z-50 md:hidden">
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-black/50"
            onClick={() => setOpen(false)}
            aria-hidden="true"
          />
          {/* Panel */}
          <aside className="fixed inset-y-0 left-0 w-72 bg-background shadow-xl">
            <div className="flex items-center justify-between border-b p-4">
              <span className="font-semibold">Menu</span>
              <button
                onClick={() => setOpen(false)}
                className="min-h-[44px] min-w-[44px] flex items-center justify-center"
                aria-label="Close menu"
              >
                ✕
              </button>
            </div>
            <nav className="overflow-y-auto p-4">{nav}</nav>
          </aside>
        </div>
      )}

      {/* Mobile header with hamburger */}
      <div className="flex flex-1 flex-col">
        <header className="flex items-center border-b p-4 md:hidden">
          <button
            onClick={() => setOpen(true)}
            className="min-h-[44px] min-w-[44px] flex items-center justify-center"
            aria-label="Open menu"
            aria-expanded={open}
          >
            <span className="sr-only">Menu</span>
            {/* Hamburger icon */}
            <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
        </header>

        <main className="flex-1 overflow-y-auto">{children}</main>
      </div>
    </div>
  );
}
```

### Bottom Tab Bar

For mobile dashboards and app-like experiences. Includes safe area support.

```tsx
import { cn } from "@/lib/utils";

interface Tab {
  label: string;
  icon: React.ReactNode;
  href: string;
}

interface BottomTabsProps {
  tabs: Tab[];
  activeHref: string;
}

export function BottomTabs({ tabs, activeHref }: BottomTabsProps) {
  return (
    <nav
      className="fixed inset-x-0 bottom-0 z-40 border-t bg-background pb-[env(safe-area-inset-bottom)] md:hidden"
      role="tablist"
    >
      <div className="flex">
        {tabs.map(tab => (
          <a
            key={tab.href}
            href={tab.href}
            role="tab"
            aria-selected={activeHref === tab.href}
            aria-current={activeHref === tab.href ? "page" : undefined}
            className={cn(
              "flex flex-1 flex-col items-center gap-1 py-2 text-xs transition-colors",
              "min-h-[44px]",
              activeHref === tab.href
                ? "text-primary"
                : "text-muted-foreground"
            )}
          >
            {tab.icon}
            <span>{tab.label}</span>
          </a>
        ))}
      </div>
    </nav>
  );
}

{/* Add spacer to prevent content from hiding behind tabs */}
{/* <div className="h-16 md:hidden" /> */}
```

### Hamburger Menu (Standalone)

Simple hamburger for content/marketing sites. Pairs with a full-screen mobile menu.

```tsx
interface HamburgerProps {
  open: boolean;
  onToggle: () => void;
}

export function HamburgerButton({ open, onToggle }: HamburgerProps) {
  return (
    <button
      onClick={onToggle}
      className="min-h-[44px] min-w-[44px] flex items-center justify-center md:hidden"
      aria-expanded={open}
      aria-label={open ? "Close menu" : "Open menu"}
    >
      <div className="relative h-5 w-6">
        <span
          className={cn(
            "absolute left-0 h-0.5 w-6 bg-foreground transition-all duration-200",
            open ? "top-2.5 rotate-45" : "top-0"
          )}
        />
        <span
          className={cn(
            "absolute left-0 top-2.5 h-0.5 w-6 bg-foreground transition-opacity duration-200",
            open ? "opacity-0" : "opacity-100"
          )}
        />
        <span
          className={cn(
            "absolute left-0 h-0.5 w-6 bg-foreground transition-all duration-200",
            open ? "top-2.5 -rotate-45" : "top-5"
          )}
        />
      </div>
    </button>
  );
}
```

### Horizontal Scrolling Tabs

For category filters or secondary navigation that overflows on mobile:

```tsx
<div className="overflow-x-auto scrollbar-none -mx-4 px-4">
  <div className="flex gap-2 whitespace-nowrap">
    {tabs.map(tab => (
      <button
        key={tab.id}
        className={cn(
          "rounded-full px-4 py-2 text-sm font-medium transition-colors",
          "min-h-[44px]",
          activeTab === tab.id
            ? "bg-primary text-primary-foreground"
            : "bg-muted text-muted-foreground"
        )}
        onClick={() => setActiveTab(tab.id)}
      >
        {tab.label}
      </button>
    ))}
  </div>
</div>
```

---

## Touch

### Touch Target Sizing

All interactive elements must be at least 44x44px. Two approaches:

**Approach 1: Padding (preferred for buttons/links)**
```tsx
{/* The padding makes the touch area large enough */}
<button className="px-4 py-3 text-sm">{label}</button>
```

**Approach 2: min-h/min-w (for icon buttons)**
```tsx
{/* Explicit minimum dimensions for small visual elements */}
<button className="min-h-[44px] min-w-[44px] flex items-center justify-center">
  <TrashIcon className="h-5 w-5" />
</button>
```

### Touch Target Spacing

Ensure enough space between targets to prevent mis-taps:

```tsx
{/* Minimum gap-2 (8px) between interactive elements */}
<div className="flex gap-2">
  <button className="min-h-[44px] px-4 py-2">Cancel</button>
  <button className="min-h-[44px] px-4 py-2">Confirm</button>
</div>
```

For vertical lists of tappable items:

```tsx
<ul className="divide-y">
  {items.map(item => (
    <li key={item.id}>
      <button className="flex w-full items-center gap-3 px-4 py-3 text-left">
        {item.label}
      </button>
    </li>
  ))}
</ul>
```

### Hover-Alternative Strategy

Desktop hover states need touch-accessible alternatives:

| Desktop pattern | Mobile alternative |
|----------------|-------------------|
| Tooltip on hover | Tap to show, or always-visible label |
| Hover to reveal actions | Always visible, or swipe to reveal, or long-press |
| Hover to preview | Tap to navigate, or show preview inline |
| Dropdown on hover | Tap to toggle dropdown |
| Hover color change | Active/tap feedback, or always show the state |

Implementation pattern:

```tsx
{/* Actions visible on mobile, hover-reveal on desktop */}
<div className="group relative">
  <div>{mainContent}</div>
  <div className="flex gap-1 md:opacity-0 md:group-hover:opacity-100 md:transition-opacity">
    <button className="min-h-[44px] min-w-[44px]">Edit</button>
    <button className="min-h-[44px] min-w-[44px]">Delete</button>
  </div>
</div>
```

### Tap Feedback

Provide visual feedback on touch for interactive elements:

```tsx
{/* Scale down slightly on press */}
<button className="transition-transform active:scale-95">
  {label}
</button>

{/* Background highlight on press */}
<button className="transition-colors active:bg-muted">
  {label}
</button>
```

---

## Forms

### Input Types and Input Modes

Using the correct `type` and `inputMode` triggers the right mobile keyboard:

| Data | `type` | `inputMode` | Keyboard shown |
|------|--------|-------------|----------------|
| Email | `email` | — | @ and .com visible |
| Phone | `tel` | — | Numeric dialpad |
| Integer | `text` | `numeric` | Number pad |
| Decimal/amount | `text` | `decimal` | Number pad with . |
| Search | `search` | — | Search action key |
| URL | `url` | — | .com and / visible |
| Password | `password` | — | Standard with hide/show |
| Date | `date` | — | Native date picker |

```tsx
{/* Amount input with decimal keyboard */}
<input
  type="text"
  inputMode="decimal"
  placeholder="0.00"
  className="w-full rounded-lg border px-4 py-3 text-base"
/>

{/* Email with correct keyboard */}
<input
  type="email"
  autoComplete="email"
  className="w-full rounded-lg border px-4 py-3 text-base"
/>
```

### Autocomplete Attributes

Help mobile browsers autofill correctly:

```tsx
{/* Full name */}
<input type="text" autoComplete="name" />

{/* Email */}
<input type="email" autoComplete="email" />

{/* Phone */}
<input type="tel" autoComplete="tel" />

{/* Street address */}
<input type="text" autoComplete="street-address" />

{/* City */}
<input type="text" autoComplete="address-level2" />

{/* Credit card */}
<input type="text" inputMode="numeric" autoComplete="cc-number" />

{/* One-time code (SMS) */}
<input type="text" inputMode="numeric" autoComplete="one-time-code" />
```

### Label Placement

Always place labels above inputs on mobile — not inline/floating:

```tsx
{/* ✅ Label above input */}
<div className="space-y-1.5">
  <label htmlFor="email" className="text-sm font-medium">
    Email address
  </label>
  <input
    id="email"
    type="email"
    autoComplete="email"
    className="w-full rounded-lg border px-4 py-3 text-base"
    placeholder="you@example.com"
  />
</div>
```

Why not floating labels:
- They shrink to tiny text on focus, reducing readability
- They depend on placeholder styling hacks
- They confuse screen readers when implemented poorly
- Label above is simpler, more accessible, and works at every size

### Mobile Form Layout

```tsx
<form className="space-y-4">
  {/* Stack all fields vertically on mobile */}
  <div className="space-y-4 sm:grid sm:grid-cols-2 sm:gap-4 sm:space-y-0">
    <div className="space-y-1.5">
      <label htmlFor="first" className="text-sm font-medium">First name</label>
      <input id="first" type="text" autoComplete="given-name"
        className="w-full rounded-lg border px-4 py-3 text-base" />
    </div>
    <div className="space-y-1.5">
      <label htmlFor="last" className="text-sm font-medium">Last name</label>
      <input id="last" type="text" autoComplete="family-name"
        className="w-full rounded-lg border px-4 py-3 text-base" />
    </div>
  </div>

  {/* Full-width submit button on mobile */}
  <button
    type="submit"
    className="w-full rounded-lg bg-primary px-4 py-3 font-medium text-primary-foreground sm:w-auto"
  >
    Submit
  </button>
</form>
```

> Use `text-base` (16px) minimum for input font size on mobile — iOS Safari zooms in on inputs smaller than 16px.
