# Layout Patterns Reference

Concrete Tailwind CSS + React patterns for responsive layouts. All examples are mobile-first.

## Responsive Container

Standard content container with responsive horizontal padding:

```tsx
<div className="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8">
  {children}
</div>
```

Narrow container for reading content:

```tsx
<div className="mx-auto w-full max-w-prose px-4 sm:px-6">
  {children}
</div>
```

## Responsive Grids

### 1→2 Columns

```tsx
<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-6">
  {items.map(item => <Card key={item.id} {...item} />)}
</div>
```

### 1→2→3 Columns

```tsx
<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 lg:gap-6">
  {items.map(item => <Card key={item.id} {...item} />)}
</div>
```

### 1→2→4 Columns

```tsx
<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4 lg:gap-6">
  {items.map(item => <Card key={item.id} {...item} />)}
</div>
```

### Auto-fill (fluid columns)

When exact column count doesn't matter, let CSS determine how many fit:

```tsx
<div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-4">
  {items.map(item => <Card key={item.id} {...item} />)}
</div>
```

## Sidebar + Content

Stacked on mobile, side-by-side from `md` up:

```tsx
<div className="flex flex-col md:flex-row md:gap-8">
  {/* Sidebar */}
  <aside className="w-full md:w-64 lg:w-72 shrink-0">
    <nav className="sticky top-4">{sidebarContent}</nav>
  </aside>

  {/* Main content */}
  <main className="min-w-0 flex-1">{children}</main>
</div>
```

With collapsible sidebar on mobile using a drawer:

```tsx
<div className="relative flex">
  {/* Mobile overlay sidebar */}
  <aside
    className={cn(
      "fixed inset-y-0 left-0 z-40 w-72 bg-background shadow-lg transition-transform md:static md:z-auto md:shadow-none",
      isOpen ? "translate-x-0" : "-translate-x-full md:translate-x-0"
    )}
  >
    {sidebarContent}
  </aside>

  {/* Overlay backdrop (mobile only) */}
  {isOpen && (
    <div
      className="fixed inset-0 z-30 bg-black/50 md:hidden"
      onClick={() => setIsOpen(false)}
    />
  )}

  <main className="min-w-0 flex-1">{children}</main>
</div>
```

## Content Stacking Order

Control visual order independent of DOM order using `order-` classes:

```tsx
{/* On mobile: image first, then content. On desktop: content left, image right */}
<div className="flex flex-col md:flex-row gap-6">
  <div className="order-2 md:order-1 flex-1">{textContent}</div>
  <div className="order-1 md:order-2 flex-1">{image}</div>
</div>
```

## Full-Width Mobile Cards

Cards that bleed to screen edges on mobile but have rounded corners on larger screens:

```tsx
<div className="-mx-4 sm:mx-0 sm:rounded-lg sm:border sm:shadow-sm">
  <div className="border-b px-4 py-3 sm:px-6">{header}</div>
  <div className="px-4 py-4 sm:px-6">{content}</div>
</div>
```

For a list of full-width mobile cards:

```tsx
<div className="divide-y -mx-4 sm:mx-0 sm:rounded-lg sm:border sm:shadow-sm sm:overflow-hidden">
  {items.map(item => (
    <div key={item.id} className="px-4 py-3 sm:px-6">
      {item.content}
    </div>
  ))}
</div>
```

## Responsive Spacing

Guidelines for spacing that adapts to screen size:

| Element | Mobile | Tablet (md) | Desktop (lg+) |
|---------|--------|-------------|----------------|
| Page padding | `px-4` | `px-6` | `px-8` |
| Section gap | `py-8` | `py-12` | `py-16` |
| Card padding | `p-4` | `p-6` | `p-6` |
| Grid gap | `gap-4` | `gap-6` | `gap-6` or `gap-8` |
| Stack gap | `gap-3` | `gap-4` | `gap-4` |

Example section with responsive spacing:

```tsx
<section className="py-8 md:py-12 lg:py-16">
  <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
    <h2 className="mb-6 md:mb-8">{title}</h2>
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-6 lg:grid-cols-3">
      {items}
    </div>
  </div>
</section>
```

## Responsive Typography

Heading sizes that scale with screen size:

```tsx
{/* Page title */}
<h1 className="text-2xl font-bold sm:text-3xl lg:text-4xl">{title}</h1>

{/* Section heading */}
<h2 className="text-xl font-semibold sm:text-2xl">{heading}</h2>

{/* Card heading */}
<h3 className="text-lg font-medium">{subheading}</h3>
```

Body text guidelines:
- `text-sm` (14px) — compact UI, secondary info, captions
- `text-base` (16px) — default body text, minimum for mobile readability
- `text-lg` (18px) — emphasized body text, hero descriptions

> Never use `text-xs` (12px) for primary content on mobile — it's too small for comfortable reading.

## Sticky Mobile CTA Bar

Fixed call-to-action at the bottom of the screen with safe area support:

```tsx
<div className="fixed inset-x-0 bottom-0 z-40 border-t bg-background p-4 pb-[max(1rem,env(safe-area-inset-bottom))] md:hidden">
  <button className="w-full rounded-lg bg-primary px-4 py-3 text-center font-medium text-primary-foreground">
    {ctaLabel}
  </button>
</div>

{/* Add matching spacer to prevent content from hiding behind CTA */}
<div className="h-20 md:hidden" />
```

## Responsive Table → Card Pattern

Tables that become stacked cards on mobile:

```tsx
{/* Desktop: table */}
<table className="hidden md:table w-full">
  <thead>
    <tr>
      <th className="text-left p-3">Name</th>
      <th className="text-left p-3">Status</th>
      <th className="text-left p-3">Amount</th>
    </tr>
  </thead>
  <tbody>
    {rows.map(row => (
      <tr key={row.id} className="border-t">
        <td className="p-3">{row.name}</td>
        <td className="p-3">{row.status}</td>
        <td className="p-3">{row.amount}</td>
      </tr>
    ))}
  </tbody>
</table>

{/* Mobile: stacked cards */}
<div className="space-y-3 md:hidden">
  {rows.map(row => (
    <div key={row.id} className="rounded-lg border p-4 space-y-2">
      <div className="font-medium">{row.name}</div>
      <div className="flex justify-between text-sm text-muted-foreground">
        <span>{row.status}</span>
        <span>{row.amount}</span>
      </div>
    </div>
  ))}
</div>
```
