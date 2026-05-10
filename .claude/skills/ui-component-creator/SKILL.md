---
name: ui-component-creator
description: Create React UI components following a project's design system. Use when building new components, buttons, cards, forms, inputs, modals, badges, or any frontend UI elements. Provides the structural patterns — project-specific skills provide the tokens.
---

# UI Component Creator

Generic skill for building React/TypeScript UI components. Defines **how** to structure components — not which colors or fonts to use.

## Before You Start

1. **Find the project's design system.** Look for `design-system.md` in the project root or standards folder. This defines the tokens (colors, fonts, spacing) to use.
2. **Check for a project-specific component skill.** Projects may have their own component skill with concrete tokens and examples. If one exists, use it together with this skill.
3. **For responsive/mobile work**, also check the `mobile-friendly-design` skill for touch targets, responsive patterns, and mobile navigation.
4. **Look at existing components** in the project to match established patterns.

If no design system exists yet, use the `design-system` skill to create one first.

## Component Structure

All components MUST use this pattern:

```tsx
import { HTMLAttributes, ReactNode, forwardRef } from "react";
import { cn } from "@/lib/utils";

export type ComponentVariant = "default" | "primary" | "accent";
export type ComponentSize = "sm" | "md" | "lg";

export interface ComponentProps extends HTMLAttributes<HTMLDivElement> {
  variant?: ComponentVariant;
  size?: ComponentSize;
  children: ReactNode;
}

const ComponentName = forwardRef<HTMLDivElement, ComponentProps>(
  ({ variant = "default", size = "md", className, children, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(baseStyles, variantStyles[variant], sizeStyles[size], className)}
        {...props}
      >
        {children}
      </div>
    );
  }
);

ComponentName.displayName = "ComponentName";
export default ComponentName;
```

## Required Patterns

### 1. Always use `forwardRef`
Every component must forward refs for parent access and composition.

### 2. Always use `cn()` for class merging
Use the `cn()` utility (typically from `@/lib/utils`) to merge base styles, variant styles, and the consumer's `className` prop.

### 3. Define variants and sizes as TypeScript types
```tsx
export type ButtonVariant = "primary" | "secondary" | "ghost" | "destructive";
export type ButtonSize = "sm" | "md" | "lg";
```

Use `Record<Variant, string>` for variant-to-class mappings:
```tsx
const variantStyles: Record<ButtonVariant, string> = {
  primary: "bg-primary text-primary-foreground",
  secondary: "bg-secondary text-secondary-foreground border border-border",
  ghost: "text-foreground hover:bg-muted",
  destructive: "bg-destructive text-destructive-foreground",
};
```

### 4. Extend native HTML attributes
```tsx
// For div-based components
interface CardProps extends HTMLAttributes<HTMLDivElement> {}

// For button-based components
interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {}

// For input-based components
interface InputProps extends InputHTMLAttributes<HTMLInputElement> {}
```

### 5. Set `displayName` and export both default and named
```tsx
ComponentName.displayName = "ComponentName";
export default ComponentName;
export { ComponentName };
```

## Semantic Color Classes

**NEVER hardcode colors.** Use the semantic classes defined in the project's design system:

| Element | Typical classes |
|---------|----------------|
| Page background | `bg-background` |
| Card background | `bg-card` / `text-card-foreground` |
| Primary action | `bg-primary` / `text-primary-foreground` |
| Secondary action | `bg-secondary` / `text-secondary-foreground` |
| Primary text | `text-foreground` |
| Muted text | `text-muted-foreground` |
| Borders | `border-border` |
| Error/destructive | `text-destructive` / `border-destructive` |
| Success | `text-success` |

> Check the project's `design-system.md` or `tailwind.config` for the actual tokens available.

## Interaction States

Every interactive component MUST include these states:

```tsx
// Hover
"hover:bg-primary-hover hover:shadow-md hover:-translate-y-0.5"

// Active
"active:translate-y-0 active:shadow-sm"

// Focus (REQUIRED for all interactive elements)
"focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"

// Disabled
"disabled:pointer-events-none disabled:opacity-50"

// Error (inputs)
"border-destructive focus:ring-destructive/20"
```

## Transitions

```tsx
"transition-all duration-normal ease-out"  // Standard (250ms)
"transition-colors duration-fast"          // Color only (150ms)
```

## Accessibility (Required)

- Use `<button>` for clickable actions, not `<div onClick>`
- Include `aria-label` for icon-only buttons
- Use `role="status"` for status indicators
- Use `sr-only` class for screen-reader-only content
- Never use `focus:outline-none` without a ring replacement

For comprehensive WCAG patterns, see the **accessibility** skill.

## Don'ts

```tsx
// ❌ WRONG
<div className="bg-[#7B2D42]">        // Hardcoded color
<div className="p-[18px]">            // Hardcoded spacing
<div onClick={handleClick}>           // Non-semantic element
"focus:outline-none"                  // Removes focus without replacement
style={{ color: 'red' }}              // Inline styles

// ✅ CORRECT
<div className="bg-primary">          // Semantic token
<div className="p-4">                 // Design system spacing
<button onClick={handleClick}>        // Semantic element
"focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
className="text-destructive"          // Semantic color
```

## Checklist Before Creating

1. Find and read the project's `design-system.md`
2. Check for a project-specific component skill
3. Look at existing components for established patterns
4. Use `forwardRef` pattern
5. Use `cn()` utility for class merging
6. Define variant and size types
7. Add focus-visible states for accessibility
8. Use semantic color classes (never hardcode)
9. Test in both light and dark mode

## How This Skill Relates to Others

| Skill | Role |
|-------|------|
| **design-system** | Creates the `design-system.md` with tokens (colors, fonts, spacing). Run this first if none exists. |
| **ui-component-creator** (this skill) | Defines component structure patterns. Framework-level: forwardRef, cn(), variants, a11y. |
| **Project-specific skill** | Provides concrete tokens, class combinations, and component examples for a specific project. |
| **front-end-design** | Creative direction for visual distinctiveness. Guides aesthetic choices, not component structure. |
| **accessibility** | WCAG compliance — semantic HTML, ARIA, keyboard nav, contrast, screen readers. |
