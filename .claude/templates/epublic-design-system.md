# Design System — ePublic Solutions

> **Source:** [epublic-solutions.nl](https://www.epublic-solutions.nl)
> **Generated:** 2026-03-04
> **Last Updated:** 2026-03-04

## Overview

The ePublic Solutions design system is rooted in a **professional, trustworthy, and modern** aesthetic befitting a public-sector technology consultancy. The visual identity centers on a deep teal/dark-cyan brand color, paired with a bright cyan accent — evoking innovation, reliability, and approachability. The overall feel is clean and spacious with generous whitespace, soft card shadows, and pill-shaped CTAs.

**Design Principles:**
1. **Trust through restraint** — limited color palette, generous whitespace, no visual clutter
2. **Accessible and inclusive** — high contrast, readable typography, WCAG-conscious design
3. **Modern yet professional** — rounded pill buttons and soft shadows balance warmth with authority

---

## Colors

### Brand Colors

| Token | HSL Value | Hex | Usage |
|-------|-----------|-----|-------|
| `primary` | `174 92% 9%` | `#022D2A` | Hero backgrounds, dark sections, footer, primary text on light |
| `primary-hover` | `174 92% 12%` | `#033D39` | Hover states on dark backgrounds |
| `primary-active` | `174 92% 7%` | `#012120` | Active/pressed on dark backgrounds |
| `accent` | `184 98% 37%` | `#01AABB` | Primary CTA buttons, links, highlights, active indicators |
| `accent-hover` | `184 98% 32%` | `#0193A0` | Hover state for accent elements |
| `accent-active` | `184 98% 27%` | `#017D88` | Active/pressed accent state |

### Neutral Colors

| Token | HSL Value | Hex | Usage |
|-------|-----------|-----|-------|
| `background` | `0 0% 97%` | `#F8F8F8` | Page background, light sections |
| `surface` | `0 0% 100%` | `#FFFFFF` | Cards, panels, elevated containers |
| `surface-elevated` | `0 0% 100%` | `#FFFFFF` | Modals, dropdowns, popovers |
| `border` | `220 13% 91%` | `#E5E7EB` | Default borders, dividers |
| `border-strong` | `218 11% 65%` | `#9CA3AF` | Emphasized borders, focused inputs |

### Text Colors

| Token | HSL Value | Hex | Usage |
|-------|-----------|-----|-------|
| `text-primary` | `174 92% 9%` | `#022D2A` | Body text, headings on light backgrounds |
| `text-heading` | `222 47% 11%` | `#111827` | Headings (used in some contexts) |
| `text-secondary` | `220 9% 46%` | `#6B7280` | Descriptions, helper text |
| `text-muted` | `218 11% 65%` | `#9CA3AF` | Placeholders, disabled text |
| `text-inverse` | `0 0% 100%` | `#FFFFFF` | Text on dark/primary backgrounds |
| `text-accent` | `184 98% 37%` | `#01AABB` | Links, labels on dark backgrounds |

### Semantic Colors

| Token | HSL Value | Hex | Usage |
|-------|-----------|-----|-------|
| `success` | `160 84% 36%` | `#047857` | Success states, positive actions |
| `success-bg` | `149 80% 90%` | `#D1FAE5` | Success alert backgrounds |
| `warning` | `38 92% 50%` | `#FBB F24` | Warnings, caution states |
| `error` | `0 84% 60%` | `#EF4444` | Errors, destructive actions |
| `error-bg` | `0 86% 97%` | `#FEF2F2` | Error alert backgrounds |
| `info` | `199 89% 48%` | `#0EA5E9` | Informational states |
| `info-bg` | `204 94% 94%` | `#E0F2FE` | Info alert backgrounds |

### Grey Scale

```
grey-50:  #F9FAFB   (lightest — backgrounds)
grey-100: #F3F4F6   (table headers, subtle backgrounds)
grey-200: #E5E7EB   (borders, dividers)
grey-300: #D1D5DB   (input borders, disabled borders)
grey-400: #9CA3AF   (muted text, placeholders)
grey-500: #6B7280   (secondary text)
grey-600: #4B5563   (body text alternative)
grey-700: #374151   (form text, strong secondary)
grey-800: #1F2937   (headings alternative)
grey-900: #111827   (darkest — primary headings)
```

---

## Typography

### Font Families

| Token | Value | Usage |
|-------|-------|-------|
| `font-display` | `'DM Sans', sans-serif` | Headings, hero text, buttons |
| `font-body` | `'DM Sans', sans-serif` | Body text, UI elements |
| `font-mono` | `'JetBrains Mono', 'Fira Code', monospace` | Code, technical content |

**Font loading:** Google Fonts — `DM Sans` weights 300, 400, 500, 600, 700.

### Font Sizes (Modular Scale — ratio 1.250)

| Token | Size | Line Height | Usage |
|-------|------|-------------|-------|
| `text-xs` | 13px / 0.8125rem | 1.5 | Badges, fine print, labels |
| `text-sm` | 14px / 0.875rem | 1.5 | Captions, helper text |
| `text-base` | 16px / 1rem | 1.5 | Base size, form labels |
| `text-body` | 18px / 1.125rem | 1.4 | Body text (default) |
| `text-lg` | 20px / 1.25rem | 1.4 | Lead paragraphs, H5 |
| `text-xl` | 25px / 1.5625rem | 1.3 | H4, card titles |
| `text-2xl` | 31.25px / 1.953rem | 1.3 | H3 |
| `text-3xl` | 39px / 2.441rem | 1.2 | H2 |
| `text-4xl` | 48.8px / 3.052rem | 1.2 | H1 (desktop hero) |
| `text-5xl` | 61px / 3.815rem | 1.1 | Display, oversized hero |

### Font Weights

| Token | Value | Usage |
|-------|-------|-------|
| `font-light` | 300 | Hero headings (H1), display text — signature ePublic style |
| `font-normal` | 400 | Body text |
| `font-medium` | 500 | Buttons, nav links, H2–H4, form labels |
| `font-semibold` | 600 | Strong emphasis, badges |
| `font-bold` | 700 | Rare — used sparingly for maximum emphasis |

### Heading Styles

| Element | Size (Desktop) | Weight | Color | Notes |
|---------|----------------|--------|-------|-------|
| H1 | text-4xl (~49px) | light (300) | text-inverse on dark, text-heading on light | Signature lightweight hero style |
| H2 | text-3xl (~39px) | medium (500) | text-primary | Section headings |
| H3 | text-2xl (~31px) | medium (500) | text-heading | Card titles, sub-sections |
| H4 | text-xl (~25px) | medium (500) | text-heading | Component titles, team names |
| H5 | text-lg (20px) | medium (500) | text-primary | Small headings, labels |
| H6 | text-base (16px) | medium (500) | text-secondary | Overlines, categories |

> **Note:** H1 uses `font-weight: 300` (light) — this is a distinctive part of the ePublic visual identity. Do not make H1 bold.

---

## Spacing

Based on a **4px** grid system, with section padding using larger values.

| Token | Value | Pixels | Common Usage |
|-------|-------|--------|--------------|
| `space-0` | 0 | 0px | Reset |
| `space-0.5` | 0.125rem | 2px | Tight inline spacing |
| `space-1` | 0.25rem | 4px | Icon padding, tight gaps |
| `space-2` | 0.5rem | 8px | Input padding, small gaps, form label spacing |
| `space-3` | 0.75rem | 12px | Button padding (sm), card gaps |
| `space-4` | 1rem | 16px | Form gaps, button padding |
| `space-5` | 1.25rem | 20px | Card padding, article padding |
| `space-6` | 1.5rem | 24px | Button padding (lg), component gaps |
| `space-8` | 2rem | 32px | Column gap, section internal spacing |
| `space-10` | 2.5rem | 40px | Large section spacing |
| `space-12` | 3rem | 48px | Section margins |
| `space-16` | 4rem | 64px | Page section vertical spacing |
| `space-24` | 6rem | 96px | Hero spacing |
| `space-25` | 6.25rem | 100px | Section vertical padding (BDE default) |

### Spacing Guidelines

- **Inline elements** (icon + text): `space-2` (8px)
- **Form fields** (label to input): `space-2` (8px)
- **Card padding**: `space-5` (20px)
- **Column gap**: `space-8` (32px)
- **Section padding**: `space-25` (100px) vertical, `space-5` (20px) horizontal
- **Page max-width**: `1200px`

---

## Border Radius

| Token | Value | Usage |
|-------|-------|-------|
| `radius-none` | 0 | Sharp corners, squared elements |
| `radius-sm` | 3px / 0.1875rem | Default buttons, form inputs (BDE default) |
| `radius-md` | 8px / 0.5rem | Small cards, tags |
| `radius-lg` | 12px / 0.75rem | shadcn/ui default (`--radius`) |
| `radius-xl` | 20px / 1.25rem | Case cards, content cards |
| `radius-2xl` | 24px / 1.5rem | Large feature cards |
| `radius-full` | 9999px | Pill buttons (CTAs), avatars, badges |

### Radius Guidelines

- **Primary/Secondary buttons**: `radius-full` (9999px) — pill shape
- **Form inputs**: `radius-sm` (3px)
- **Cards**: `radius-xl` (20px)
- **Modals**: `radius-lg` (12px)
- **Avatars**: `radius-full`
- **Tags/Badges**: `radius-full`
- **Nav CTA button**: `radius-full`

---

## Shadows

| Token | Value | Usage |
|-------|-------|-------|
| `shadow-none` | none | Flat elements |
| `shadow-sm` | `0 1px 3px rgba(0,0,0,0.05), 0 1px 2px rgba(0,0,0,0.05)` | Subtle lift, form wrappers |
| `shadow-card` | `2px 4px 20px rgba(0,0,0,0.06)` | Cards at rest (signature ePublic shadow) |
| `shadow-md` | `0 4px 6px rgba(0,0,0,0.1)` | Hovered cards, raised buttons |
| `shadow-lg` | `0 10px 15px rgba(0,0,0,0.1)` | Dropdowns, popovers |
| `shadow-xl` | `0 20px 25px rgba(0,0,0,0.15)` | Modals |

### Elevation Guidelines

| Level | Shadow | Examples |
|-------|--------|----------|
| 0 | none | Flat UI, inline elements, dark-bg sections |
| 1 | shadow-card | Cards at rest |
| 2 | shadow-md | Hovered cards, raised buttons |
| 3 | shadow-lg | Dropdowns, tooltips, popovers |
| 4 | shadow-xl | Modals, dialogs |

---

## Breakpoints

| Token | Value | Target |
|-------|-------|--------|
| `sm` | 640px | Large phones, landscape |
| `md` | 768px | Tablets |
| `lg` | 1024px | Small laptops |
| `xl` | 1280px | Laptops, desktops |
| `2xl` | 1440px | Large desktops |

### Responsive Strategy

**Mobile-first approach.** Default styles = mobile, add complexity at larger breakpoints.

- Max content width: `1200px` (centered)
- Container padding: `20px` (mobile) → `2rem` (desktop)
- Touch targets: minimum 44x44px on mobile
- Navigation: hamburger menu below `lg`, horizontal nav at `lg`+

---

## Component Patterns

### Buttons

#### Primary Button (Pill CTA)

```
Background: accent (#01AABB)
Text: text-inverse (white)
Padding: 15px 30px
Border Radius: radius-full (9999px)
Font: font-body (DM Sans), text-body (18px), font-medium (500)
Border: none
Shadow: none

Hover: accent-hover (#0193A0)
Active: accent-active (#017D88)
Disabled: opacity 50%, cursor not-allowed
Transition: 300ms
```

#### Secondary Button (Outline Pill)

```
Background: transparent
Border: 1px solid white (on dark) / 1px solid border (on light)
Text: text-inverse (on dark) / text-primary (on light)
Padding: 15px 30px
Border Radius: radius-full (9999px)
Font: font-body (DM Sans), text-body (18px), font-medium (500)

Hover: surface background at 10% opacity
Active: surface background at 20% opacity
Disabled: opacity 50%
Transition: 300ms
```

#### Ghost Button (Text Link with Arrow)

```
Background: transparent
Border: none
Text: text-accent (#01AABB) / or text-primary
Font: text-body (18px), font-medium (500)
Decoration: underline on hover

Used for: "Bekijk case" style inline links
```

#### Nav CTA Button

```
Background: accent (#01AABB)
Text: white
Padding: 15px 30px
Border Radius: radius-full (9999px)
Font: 18px, font-medium (500)

Used for: navigation bar "Kennismaken" button
```

#### Button Sizes

| Size | Padding | Font Size | Min Height |
|------|---------|-----------|------------|
| sm | 8px 16px | text-sm (14px) | 36px |
| md | 14px 24px | text-base (16px) | 44px |
| lg | 15px 30px | text-body (18px) | 50px |

---

### Cards

#### Case Card

```
Background: surface (white)
Border: none
Border Radius: radius-xl (20px)
Padding: space-5 (20px)
Shadow: shadow-card (2px 4px 20px rgba(0,0,0,0.06))

Hover: shadow-md
Transition: 300ms

Contains:
- Image (border-radius: radius-xl on top or full)
- H3 title (text-2xl, font-medium)
- Description text (text-body, text-primary)
- "Bekijk case" link (ghost button style)
```

#### Team Member Card

```
Background: none (transparent on carousel)
Border: none
Padding: space-4
Text Align: center

Contains:
- Avatar image (rounded, large)
- Name (H4, text-xl, font-medium)
- Role (text-body, text-secondary)
```

#### Testimonial Card

```
Background: surface (white)
Border: none
Border Radius: radius-xl (20px)
Padding: space-6 (24px)

Contains:
- Decorative quote image
- Quote text (text-body, italic)
- Author image + name (H4) + role
```

---

### Inputs

#### Text Input

```
Background: surface (white)
Border: 1px solid grey-300 (#D1D5DB)
Border Radius: radius-sm (3px)
Padding: 12px 16px
Font: text-base (16px)
Text Color: grey-700 (#374151)
Placeholder Color: grey-450 (#787E8B)

Focus: border-color accent, ring (0 0 0 2px accent/20%)
Error: border-color error (#EF4444)
Disabled: background grey-100, cursor not-allowed
```

#### Input with Label

```
Label: text-sm or text-base, font-medium (500)
Gap (label to input): space-2 (8px)
Gap (input to helper): space-2 (8px)
Gap (between form fields): space-4 (16px)
```

---

### Sections

#### Dark Section (Hero / CTA)

```
Background: primary (#022D2A)
Text: text-inverse (white)
Subtitle: text-accent (#01AABB) or white at reduced opacity
Padding: space-25 (100px) vertical

Used for: Hero, "Trusted advisors", footer, CTA blocks
```

#### Light Section

```
Background: background (#F8F8F8)
Text: text-primary (#022D2A)
Padding: space-25 (100px) vertical

Used for: Content sections, case overview, team carousel
```

#### Section with Image Overlay

```
Background: primary at 92% opacity over image
Text: text-inverse
```

---

### Navigation

#### Desktop (≥ lg)

```
Background: transparent (overlaying hero) or primary
Height: ~80px
Logo: left-aligned
Links: center, text-body (18px), font-medium (500), white
CTA: right-aligned, pill button (accent bg, white text)

Link hover: text-accent
Active: text-accent or underline
```

#### Mobile (< lg)

```
Hamburger icon: right-aligned, white
Menu: full-screen overlay or slide-in panel
Background: primary
Links: stacked vertically, large touch targets
```

---

### Footer

```
Background: primary (#022D2A)
Text: white (headings), grey-400 (body)
Padding: space-16 vertical

Columns: 4 (logo+tagline, menu, diensten, contact)
Link style: white, no underline, underline on hover
Divider: thin border (grey-700) between content and copyright

Copyright bar: text-sm, grey-400
```

---

## Dark Mode

The ePublic website itself does not use a dark mode toggle — the design already uses dark sections (primary background) paired with light sections. For the Project Assistent app, dark mode is supported via the `.dark` class (shadcn/ui convention).

| Token | Light | Dark |
|-------|-------|------|
| background | #F8F8F8 | #0A1614 |
| surface | #FFFFFF | #0F221F |
| surface-elevated | #FFFFFF | #163330 |
| text-primary | #022D2A | #F9FAFB |
| text-secondary | #6B7280 | #9CA3AF |
| text-muted | #9CA3AF | #6B7280 |
| border | #E5E7EB | #1F3D39 |
| accent | #01AABB | #01AABB |

---

## Transition & Animation

| Token | Value | Usage |
|-------|-------|-------|
| `transition-default` | `300ms ease` | All interactive state changes |
| `transition-fast` | `150ms ease` | Tooltips, micro-interactions |
| `transition-slow` | `500ms ease` | Page transitions, carousels |

---

## Implementation Notes

### CSS Variables (shadcn/ui-compatible HSL format)

```css
:root {
  /* Brand */
  --primary: 174 92% 9%;          /* #022D2A — deep teal */
  --primary-foreground: 0 0% 100%; /* white */
  --accent: 184 98% 37%;           /* #01AABB — bright cyan */
  --accent-foreground: 0 0% 100%;  /* white */

  /* Surfaces */
  --background: 0 0% 97%;          /* #F8F8F8 */
  --foreground: 174 92% 9%;        /* #022D2A */
  --card: 0 0% 100%;               /* white */
  --card-foreground: 174 92% 9%;   /* #022D2A */
  --popover: 0 0% 100%;
  --popover-foreground: 222 47% 11%;

  /* Secondary */
  --secondary: 220 14% 96%;
  --secondary-foreground: 174 92% 9%;

  /* Muted */
  --muted: 220 14% 96%;
  --muted-foreground: 220 9% 46%;

  /* Destructive */
  --destructive: 0 84% 60%;
  --destructive-foreground: 0 0% 100%;

  /* Borders */
  --border: 220 13% 91%;
  --input: 218 11% 75%;
  --ring: 184 98% 37%;             /* accent for focus rings */

  /* Radius */
  --radius: 0.75rem;

  /* Semantic */
  --success: 160 84% 36%;
  --success-foreground: 0 0% 100%;
  --warning: 38 92% 50%;
  --warning-foreground: 0 0% 100%;
  --info: 199 89% 48%;
  --info-foreground: 0 0% 100%;

  /* Sidebar */
  --sidebar-background: 174 92% 9%;       /* dark teal sidebar */
  --sidebar-foreground: 0 0% 100%;
  --sidebar-primary: 184 98% 37%;
  --sidebar-primary-foreground: 0 0% 100%;
  --sidebar-accent: 174 60% 14%;
  --sidebar-accent-foreground: 0 0% 100%;
  --sidebar-border: 174 40% 16%;
  --sidebar-ring: 184 98% 37%;
}
```

### Tailwind Config Additions

```typescript
// tailwind.config.ts — extend theme
{
  theme: {
    extend: {
      fontFamily: {
        display: ['DM Sans', 'sans-serif'],
        body: ['DM Sans', 'sans-serif'],
      },
      boxShadow: {
        card: '2px 4px 20px rgba(0, 0, 0, 0.06)',
      },
      maxWidth: {
        section: '1200px',
      },
    }
  }
}
```

### Google Fonts Import

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=DM+Sans:ital,opsz,wght@0,9..40,300;0,9..40,400;0,9..40,500;0,9..40,600;0,9..40,700;1,9..40,300;1,9..40,400&display=swap" rel="stylesheet">
```

---

## Changelog

| Date | Change | Author |
|------|--------|--------|
| 2026-03-04 | Initial design system extracted from epublic-solutions.nl | Claude |
