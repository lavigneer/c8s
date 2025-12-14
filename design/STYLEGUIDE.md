# CI Dashboard Design System Style Guide

## Overview

This style guide documents the complete design system for the CI/CD dashboard. It provides all the necessary specifications to reproduce the interface accurately with any development tool or agent.

---

## Color System

### Design Token Architecture

The application uses CSS custom properties (design tokens) for all colors, defined in both light and dark modes using the OKLCH color space for perceptually uniform colors.

### Light Mode Colors

\`\`\`css
:root {
  --background: oklch(1 0 0);                    /* Pure white */
  --foreground: oklch(0.145 0 0);                /* Near black */
  --card: oklch(1 0 0);                          /* White */
  --card-foreground: oklch(0.145 0 0);           /* Near black */
  --popover: oklch(1 0 0);                       /* White */
  --popover-foreground: oklch(0.145 0 0);        /* Near black */
  --primary: oklch(0.205 0 0);                   /* Dark gray/black */
  --primary-foreground: oklch(0.985 0 0);        /* Off-white */
  --secondary: oklch(0.97 0 0);                  /* Light gray */
  --secondary-foreground: oklch(0.205 0 0);      /* Dark gray */
  --muted: oklch(0.97 0 0);                      /* Light gray */
  --muted-foreground: oklch(0.556 0 0);          /* Medium gray */
  --accent: oklch(0.97 0 0);                     /* Light gray */
  --accent-foreground: oklch(0.205 0 0);         /* Dark gray */
  --destructive: oklch(0.577 0.245 27.325);      /* Red */
  --destructive-foreground: oklch(0.577 0.245 27.325); /* Red */
  --border: oklch(0.922 0 0);                    /* Light gray border */
  --input: oklch(0.922 0 0);                     /* Light gray */
  --ring: oklch(0.708 0 0);                      /* Medium gray */
  --radius: 0.625rem;                            /* 10px border radius */
}
\`\`\`

### Dark Mode Colors

\`\`\`css
.dark {
  --background: oklch(0.145 0 0);                /* Near black */
  --foreground: oklch(0.985 0 0);                /* Off-white */
  --card: oklch(0.145 0 0);                      /* Near black */
  --card-foreground: oklch(0.985 0 0);           /* Off-white */
  --popover: oklch(0.145 0 0);                   /* Near black */
  --popover-foreground: oklch(0.985 0 0);        /* Off-white */
  --primary: oklch(0.985 0 0);                   /* Off-white */
  --primary-foreground: oklch(0.205 0 0);        /* Dark gray */
  --secondary: oklch(0.269 0 0);                 /* Dark gray */
  --secondary-foreground: oklch(0.985 0 0);      /* Off-white */
  --muted: oklch(0.269 0 0);                     /* Dark gray */
  --muted-foreground: oklch(0.708 0 0);          /* Medium gray */
  --accent: oklch(0.269 0 0);                    /* Dark gray */
  --accent-foreground: oklch(0.985 0 0);         /* Off-white */
  --destructive: oklch(0.396 0.141 25.723);      /* Dark red */
  --destructive-foreground: oklch(0.637 0.237 25.331); /* Light red */
  --border: oklch(0.269 0 0);                    /* Dark gray border */
  --input: oklch(0.269 0 0);                     /* Dark gray */
  --ring: oklch(0.439 0 0);                      /* Medium dark gray */
}
\`\`\`

### Status Colors

Independent of theme, used for status indicators:

\`\`\`css
/* Success - Green */
.bg-green-500 { background: #22c55e; }
.text-green-500 { color: #22c55e; }

/* Error/Failed - Red */
.bg-red-500 { background: #ef4444; }
.text-red-500 { color: #ef4444; }

/* Running/In Progress - Blue */
.bg-blue-500 { background: #3b82f6; }
.text-blue-500 { color: #3b82f6; }

/* Commit - Purple */
.bg-purple-500 { background: #a855f7; }
.text-purple-500 { color: #a855f7; }
\`\`\`

### Semantic Usage

- **background**: Main page background
- **foreground**: Primary text color
- **card**: Card/panel backgrounds
- **card-foreground**: Text on cards
- **primary**: Primary buttons, important actions
- **primary-foreground**: Text on primary elements
- **secondary**: Secondary buttons, less prominent actions
- **secondary-foreground**: Text on secondary elements
- **muted**: Subdued backgrounds (disabled states, subtle sections)
- **muted-foreground**: Secondary text (metadata, timestamps, hints)
- **accent**: Hover states, focused elements
- **accent-foreground**: Text on accent elements
- **border**: Dividers, card borders, input borders
- **destructive**: Error states, delete actions
- **ring**: Focus ring outlines

---

## Typography

### Font Families

**Primary Font (Sans-serif)**: Geist
- Used for all UI text, headings, and body content
- Applied via `font-sans` Tailwind class
- Character set: Latin

**Monospace Font**: Geist Mono
- Used for code snippets, commit hashes, branch names, technical identifiers
- Applied via `font-mono` Tailwind class
- Character set: Latin

### Font Configuration

\`\`\`typescript
// Next.js font import
import { Geist, Geist_Mono } from 'next/font/google'

const geist = Geist({ subsets: ['latin'] })
const geistMono = Geist_Mono({ subsets: ['latin'] })
\`\`\`

### Type Scale

| Element | Size | Weight | Line Height | Tailwind Classes |
|---------|------|--------|-------------|------------------|
| Page Title (H1) | 24px (1.5rem) | 700 (Bold) | 1.2 | `text-2xl font-bold` |
| Section Title (H2) | 18px (1.125rem) | 600 (Semibold) | 1.4 | `text-lg font-semibold` |
| Card Title (H3) | 16px (1rem) | 600 (Semibold) | 1.5 | `text-base font-semibold` |
| Body Text | 14px (0.875rem) | 400 (Regular) | 1.5 | `text-sm` |
| Small Text | 12px (0.75rem) | 400 (Regular) | 1.5 | `text-xs` |
| Extra Small | 10px (0.625rem) | 400 (Regular) | 1.4 | `text-[10px]` |
| Stat Value | 24px (1.5rem) | 600 (Semibold) | 1.2 | `text-2xl font-semibold` |
| Monospace Code | 12px (0.75rem) | 400 (Regular) | 1.5 | `font-mono text-xs` |
| Monospace Branch | 14px (0.875rem) | 600 (Semibold) | 1.5 | `font-mono text-sm font-semibold` |

### Text Colors

- **Primary text**: `text-foreground`
- **Secondary/metadata text**: `text-muted-foreground`
- **Links**: `hover:text-foreground hover:underline` (starts as muted)
- **Status text on badges**: `text-white` (always white regardless of theme)

### Anti-aliasing

All text uses `antialiased` class for smooth rendering.

\`\`\`html
<body className="font-sans antialiased">
\`\`\`

---

## Spacing System

### Tailwind Spacing Scale

The design uses Tailwind's default spacing scale (1 unit = 0.25rem = 4px):

| Spacing | Rem | Pixels | Usage |
|---------|-----|--------|-------|
| `p-1` / `gap-1` | 0.25rem | 4px | Minimal spacing |
| `p-2` / `gap-2` | 0.5rem | 8px | Tight spacing, badges |
| `p-3` / `gap-3` | 0.75rem | 12px | Compact spacing, icons |
| `p-4` / `gap-4` | 1rem | 16px | Standard spacing, cards |
| `p-6` / `gap-6` | 1.5rem | 24px | Large spacing, sections |
| `p-8` / `gap-8` | 2rem | 32px | Extra large spacing |

### Component Spacing Patterns

**Card Padding**: `p-4` (16px) for standard cards, `p-6` (24px) for headers/featured cards

**Grid Gaps**: `gap-4` (16px) for stat grids, `gap-6` (24px) for major sections

**Inline Gaps**: `gap-2` (8px) for inline elements (badges + text), `gap-3` (12px) for icons + text

**Container**: `container mx-auto` with responsive padding

---

## Layout Patterns

### Layout Method Priority

1. **Flexbox** (primary): Used for most layouts
2. **CSS Grid** (secondary): Used only for stat cards and complex 2D layouts
3. **Absolute positioning**: Avoided unless necessary

### Common Layout Patterns

#### Horizontal Layout with Space Between
\`\`\`html
<div class="flex items-center justify-between">
  <!-- Content -->
</div>
\`\`\`

#### Vertical Stack
\`\`\`html
<div class="space-y-4">
  <!-- Stacked items with 16px gap -->
</div>
\`\`\`

#### Grid Layout (Stats)
\`\`\`html
<div class="grid gap-4 md:grid-cols-4">
  <!-- 4 columns on medium+ screens -->
</div>
\`\`\`

#### Centered Container
\`\`\`html
<div class="container mx-auto p-4">
  <!-- Max-width centered content -->
</div>
\`\`\`

---

## Border Radius

### Radius Scale

Defined in globals.css:

\`\`\`css
--radius: 0.625rem;  /* 10px - base radius */

/* Derived values */
--radius-sm: calc(var(--radius) - 4px);  /* 6px */
--radius-md: calc(var(--radius) - 2px);  /* 8px */
--radius-lg: var(--radius);              /* 10px */
--radius-xl: calc(var(--radius) + 4px);  /* 14px */
\`\`\`

### Usage

| Element | Class | Computed Value |
|---------|-------|----------------|
| Buttons | `rounded-lg` | 10px |
| Cards | `rounded-lg` | 10px |
| Status badges | `rounded-full` | 9999px (fully round) |
| Small badges | `rounded` | 4px |
| Icons containers | `rounded-lg` or `rounded-full` | 10px or circle |

---

## Component Specifications

### Buttons

#### Primary Button
\`\`\`html
<button class="rounded-lg bg-primary px-3 py-1 text-sm text-primary-foreground hover:bg-accent">
  Text
</button>
\`\`\`

#### Secondary Button
\`\`\`html
<button class="rounded-lg border border-border bg-secondary px-3 py-1 text-sm hover:bg-accent">
  Text
</button>
\`\`\`

#### Large Button
\`\`\`html
<button class="rounded-lg border border-border bg-secondary px-4 py-2 text-sm hover:bg-accent">
  Text
</button>
\`\`\`

#### Button with Icon
\`\`\`html
<button class="rounded-lg border border-border bg-secondary p-2 hover:bg-accent">
  <svg class="size-5"><!-- icon --></svg>
</button>
\`\`\`

### Cards

#### Standard Card
\`\`\`html
<div class="rounded-lg border border-border bg-card">
  <div class="border-b border-border p-4">
    <h2 class="text-lg font-semibold">Title</h2>
  </div>
  <div class="p-4">
    <!-- Content -->
  </div>
</div>
\`\`\`

#### Stat Card
\`\`\`html
<div class="rounded-lg border border-border bg-card p-4">
  <div class="text-sm text-muted-foreground">Label</div>
  <div class="mt-2 flex items-end justify-between">
    <div class="text-2xl font-semibold">Value</div>
    <div class="text-sm text-green-500">+12%</div>
  </div>
</div>
\`\`\`

#### Hover Card
\`\`\`html
<div class="rounded-lg border border-border bg-card p-4 hover:bg-accent/50">
  <!-- Content -->
</div>
\`\`\`

### Status Badges

#### Success Badge
\`\`\`html
<span class="inline-flex items-center gap-1 rounded-full bg-green-500 px-2 py-0.5 text-xs text-white">
  <svg class="size-4"><!-- checkmark icon --></svg>
  success
</span>
\`\`\`

#### Failed Badge
\`\`\`html
<span class="inline-flex items-center gap-1 rounded-full bg-red-500 px-2 py-0.5 text-xs text-white">
  <svg class="size-4"><!-- X icon --></svg>
  failed
</span>
\`\`\`

#### Running Badge
\`\`\`html
<span class="inline-flex items-center gap-1 rounded-full bg-blue-500 px-2 py-0.5 text-xs text-white">
  <svg class="size-4 animate-spin"><!-- spinner icon --></svg>
  running
</span>
\`\`\`

### Icon Containers

#### Small Icon Badge (Activity Feed)
\`\`\`html
<div class="flex size-8 items-center justify-center rounded-full bg-green-500/10 text-green-500">
  <svg class="size-4"><!-- icon --></svg>
</div>
\`\`\`

#### Medium Icon Badge (Stages)
\`\`\`html
<div class="flex size-8 items-center justify-center rounded-full bg-green-500 text-white">
  <svg class="size-4"><!-- icon --></svg>
</div>
\`\`\`

#### Large Icon Container (Project Header)
\`\`\`html
<div class="flex size-12 items-center justify-center rounded-lg bg-primary/10">
  <svg class="size-6 text-primary"><!-- icon --></svg>
</div>
\`\`\`

### Dividers

#### Horizontal Divider
\`\`\`html
<div class="divide-y divide-border">
  <!-- Items automatically get top borders except first -->
</div>
\`\`\`

#### Border Between Sections
\`\`\`html
<div class="border-b border-border"></div>
\`\`\`

### Lists

#### Pipeline List Item
\`\`\`html
<div class="p-4 hover:bg-accent/50">
  <div class="flex items-start justify-between gap-4">
    <div class="flex-1 space-y-1">
      <!-- Content -->
    </div>
    <!-- Action button -->
  </div>
</div>
\`\`\`

### Links

#### Breadcrumb Link
\`\`\`html
<Link href="/" class="hover:text-foreground">
  Dashboard
</Link>
\`\`\`

#### Inline Link with Icon
\`\`\`html
<Link href="/path" class="text-sm text-muted-foreground hover:text-foreground hover:underline">
  project-name
</Link>
\`\`\`

### Filter Buttons

#### Active Filter
\`\`\`html
<button class="rounded-lg bg-primary px-3 py-1 text-sm text-primary-foreground">
  All
</button>
\`\`\`

#### Inactive Filter
\`\`\`html
<button class="rounded-lg bg-secondary px-3 py-1 text-sm hover:bg-accent">
  Success
</button>
\`\`\`

---

## Icons

### Icon System

All icons are inline SVG with stroke-based design (no fills except circles).

### Icon Specifications

- **Stroke width**: `2` (standard), `1.5` for thinner icons
- **View box**: `0 0 24 24`
- **Current color**: Icons use `stroke="currentColor"` to inherit text color
- **Fill**: `fill="none"` (stroke-only design)

### Icon Sizes

| Context | Size Class | Pixels |
|---------|-----------|--------|
| Small inline | `size-3` | 12px |
| Standard inline | `size-4` | 16px |
| Medium standalone | `size-5` | 20px |
| Large header | `size-6` | 24px |
| Logo/branding | `size-8` | 32px |

### Common Icons

#### Clock Icon
\`\`\`html
<svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
  <circle cx="12" cy="12" r="10" />
  <path d="M12 6v6l4 2" />
</svg>
\`\`\`

#### Checkmark Icon
\`\`\`html
<svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
  <polyline points="20 6 9 17 4 12" />
</svg>
\`\`\`

#### X Icon (Failed)
\`\`\`html
<svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
  <line x1="18" y1="6" x2="6" y2="18" />
  <line x1="6" y1="6" x2="18" y2="18" />
</svg>
\`\`\`

#### Spinner Icon (Running)
\`\`\`html
<svg class="size-4 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
  <path d="M21 12a9 9 0 1 1-6.219-8.56" />
</svg>
\`\`\`

#### User Icon
\`\`\`html
<svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
  <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" />
  <circle cx="12" cy="7" r="4" />
</svg>
\`\`\`

#### GitHub Icon
\`\`\`html
<svg class="size-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
  <path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4" />
  <path d="M9 18c-4.51 2-5-2-7-2" />
</svg>
\`\`\`

#### Sun Icon (Light Mode)
\`\`\`html
<svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
  <circle cx="12" cy="12" r="5" />
  <line x1="12" y1="1" x2="12" y2="3" />
  <line x1="12" y1="21" x2="12" y2="23" />
  <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
  <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
  <line x1="1" y1="12" x2="3" y2="12" />
  <line x1="21" y1="12" x2="23" y2="12" />
  <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
  <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
</svg>
\`\`\`

#### Moon Icon (Dark Mode)
\`\`\`html
<svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
  <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
</svg>
\`\`\`

---

## Dark Mode Implementation

### Theme Toggle

The theme is controlled via a `.dark` class on the `<html>` element:

\`\`\`typescript
// Toggle theme
const toggleTheme = () => {
  const newIsDark = !isDark
  setIsDark(newIsDark)
  document.documentElement.classList.toggle('dark', newIsDark)
  localStorage.setItem('theme', newIsDark ? 'dark' : 'light')
}

// Initialize theme on load
useEffect(() => {
  const savedTheme = localStorage.getItem('theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  const shouldBeDark = savedTheme === 'dark' || (!savedTheme && prefersDark)
  
  setIsDark(shouldBeDark)
  document.documentElement.classList.toggle('dark', shouldBeDark)
}, [])
\`\`\`

### CSS Configuration

\`\`\`css
@custom-variant dark (&:is(.dark *));
\`\`\`

This Tailwind variant enables `.dark` class-based theme switching.

---

## Responsive Design

### Breakpoints

Tailwind default breakpoints:
- `sm`: 640px
- `md`: 768px
- `lg`: 1024px
- `xl`: 1280px
- `2xl`: 1536px

### Mobile-First Patterns

#### Stats Grid
\`\`\`html
<!-- 1 column mobile, 4 columns desktop -->
<div class="grid gap-4 md:grid-cols-4">
\`\`\`

#### Responsive Padding
\`\`\`html
<div class="container mx-auto p-4">
  <!-- 16px padding all screens -->
</div>
\`\`\`

---

## Animation

### Spinner Animation

\`\`\`html
<svg class="animate-spin">
  <!-- Rotating loading spinner -->
</svg>
\`\`\`

### Hover Transitions

All interactive elements use implicit browser transitions:
- Buttons: `hover:bg-accent`
- Cards: `hover:bg-accent/50`
- Links: `hover:text-foreground hover:underline`

No explicit transition classes needed - browser defaults are used.

---

## Accessibility

### Semantic HTML

- Use semantic elements: `<header>`, `<main>`, `<nav>`, `<section>`
- Use `<button>` for actions, `<a>` / `<Link>` for navigation
- Use heading hierarchy (`h1`, `h2`, `h3`)

### ARIA Labels

\`\`\`html
<button aria-label="Toggle theme">
  <!-- Icon only button needs label -->
</button>
\`\`\`

### Focus States

All interactive elements have focus outlines:

\`\`\`css
* {
  @apply outline-ring/50;
}
\`\`\`

### Color Contrast

All color combinations meet WCAG AA standards:
- Light mode: Dark text on light backgrounds
- Dark mode: Light text on dark backgrounds
- Status badges: Always white text on colored backgrounds

---

## Page Layouts

### Dashboard Layout
\`\`\`
- Header (fixed, full width)
  - Logo + Title (left)
  - Theme toggle (right)
- Main content (container, centered)
  - Stats grid (4 columns)
  - Projects list card
  - Recent builds card
  - Activity feed card
\`\`\`

### Project Detail Layout
\`\`\`
- Header (same as dashboard)
- Main content (container, centered)
  - Breadcrumb
  - Project info card (with icon, metadata, settings button)
  - Pipeline runs card (with filter buttons)
\`\`\`

### Pipeline Detail Layout
\`\`\`
- Header (same as dashboard)
- Main content (container, centered)
  - Breadcrumb
  - Pipeline header card (with status, metadata, action buttons)
  - Stages list (expandable cards)
  - Artifacts card
\`\`\`

---

## Data Display Patterns

### Metadata Display

#### Inline Metadata with Icons
\`\`\`html
<div class="flex items-center gap-4 text-xs text-muted-foreground">
  <span class="flex items-center gap-1">
    <svg class="size-3"><!-- icon --></svg>
    john.doe
  </span>
  <span>•</span>
  <span class="flex items-center gap-1">
    <svg class="size-3"><!-- icon --></svg>
    4m 12s
  </span>
  <span>•</span>
  <span>2 minutes ago</span>
</div>
\`\`\`

### Commit Hash Display
\`\`\`html
<span class="rounded bg-secondary px-2 py-0.5 font-mono text-xs text-muted-foreground">
  a3f82e1
</span>
\`\`\`

### Branch Name Display
\`\`\`html
<span class="font-mono text-sm font-semibold">main</span>
\`\`\`

### Percentage Change Display
\`\`\`html
<!-- Positive -->
<div class="text-sm text-green-500">+12%</div>

<!-- Negative -->
<div class="text-sm text-red-500">-3%</div>
\`\`\`

---

## File Structure

### Required Files

\`\`\`
app/
├── globals.css          # Design tokens, Tailwind config
├── layout.tsx           # Root layout with fonts
├── page.tsx            # Dashboard page
├── project/
│   └── [id]/
│       ├── page.tsx    # Project detail page
│       └── pipeline/
│           └── [id]/
│               └── page.tsx  # Pipeline detail page
components/
├── dashboard-header.tsx      # Header with theme toggle
├── builds-overview.tsx       # Stats cards
├── pipelines-list.tsx        # Recent builds list
├── recent-activity.tsx       # Activity feed
├── projects-list.tsx         # Projects grid
├── project-detail.tsx        # Project detail component
└── pipeline-detail.tsx       # Pipeline detail component
\`\`\`

---

## Implementation Checklist

When reproducing this design:

- [ ] Set up CSS custom properties for all color tokens (light + dark)
- [ ] Configure Tailwind with custom variant for dark mode
- [ ] Import and configure Geist and Geist Mono fonts
- [ ] Apply `font-sans antialiased` to body
- [ ] Implement theme toggle with localStorage persistence
- [ ] Use semantic HTML elements
- [ ] Apply proper ARIA labels to icon-only buttons
- [ ] Use design tokens (not hardcoded colors) for all styling
- [ ] Follow spacing scale (prefer gap over margin/padding when possible)
- [ ] Use flexbox as primary layout method
- [ ] Ensure all icons are stroke-based SVG with `currentColor`
- [ ] Test color contrast in both light and dark modes
- [ ] Verify responsive behavior on mobile and desktop
- [ ] Add hover states to all interactive elements
- [ ] Use proper border-radius values (lg = 10px, full for badges)
- [ ] Apply status colors consistently (green/red/blue/purple)
- [ ] Use monospace font for technical identifiers
- [ ] Implement breadcrumb navigation on detail pages
- [ ] Add proper link styling with hover underlines
- [ ] Test keyboard navigation and focus states

---

## Technical Stack

- **Framework**: Next.js (App Router)
- **Styling**: Tailwind CSS v4
- **Fonts**: Geist (sans), Geist Mono (mono) from Google Fonts
- **Icons**: Inline SVG (stroke-based)
- **Theme**: CSS custom properties with class-based dark mode
- **State**: React hooks (useState, useEffect)
- **Routing**: Next.js file-based routing with dynamic segments
- **Color Space**: OKLCH for perceptually uniform colors

---

## Notes for Reproduction

1. **Do not use Radix UI or any component library** - This design is pure HTML + Tailwind CSS
2. **Always use design tokens** - Never hardcode colors like `bg-white` or `text-black`
3. **Status colors are exceptions** - Green/red/blue/purple are hardcoded for status indicators
4. **Maintain consistent spacing** - Use Tailwind's spacing scale, prefer `gap` over `margin`
5. **Icons must match exactly** - All icons are stroke-based, no fills
6. **Hover states are critical** - All interactive elements need hover styling
7. **Dark mode is required** - Implement full theme toggle with persistence
8. **Use semantic HTML** - Proper heading hierarchy and element types
9. **Mobile-first approach** - Start with mobile layout, enhance for desktop
10. **Accessibility matters** - ARIA labels, focus states, color contrast all required

---

This style guide contains all specifications needed to reproduce the CI dashboard design pixel-perfectly in any development environment or with any AI agent.
