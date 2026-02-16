# Dark Mode Implementation Plan

## Step 1: Clarify What We're Building

Implement dark mode as the default theme for the Nuxt/Vue frontend with blue accents on dark background.

**Scope:**
- Add dark mode support to existing Tailwind CSS setup
- Make dark mode the default
- Blue accent colors (primary-500 and related shades) on dark backgrounds
- No light mode alternative needed initially

**Constraints:**
- Must use existing Tailwind setup
- Primary blue colors already defined in tailwind.config.js
- No visual mocks provided

## Step 2: Visuals

No visuals provided.

## Step 3: Reference Implementations

No similar dark mode code in this codebase to reference.

## Step 4: Product Context

Product mission focuses on inventory and listing management dashboard for personal use. Dark mode improves usability for extended usage sessions.

## Step 5: Standards

No specific standards apply to UI/theming. General tech stack from index.yml:
- Tech stack conventions (use established patterns)

## Step 6: Spec Folder Name

`2026-02-03-1200-dark-mode-frontend/`

## Step 7: Plan Structure

```
agent-os/specs/2026-02-03-1200-dark-mode-frontend/
├── plan.md
├── shape.md
├── standards.md
└── references.md
```

## Step 8: Implementation Tasks

### Task 1: Save spec documentation ✅ DONE

### Task 2: Configure Tailwind CSS for dark mode ✅ DONE
- Enable `darkMode: 'class'` in tailwind.config.js
- Extend colors for dark mode (dark backgrounds, adjusted text colors)
- Update component classes with dark: variants

### Task 3: Create theme composable ✅ DONE (NOT NEEDED)
- Dark mode is applied by default via `class="dark"` in app.vue
- No toggle needed for MVP

### Task 4: Update main.css with dark mode base styles ✅ DONE
- Change body background to dark
- Set text colors for dark mode
- Define dark variants for all component classes

### Task 5: Update layouts for theme support ✅ DONE
- Apply dark class to html/root element
- Ensure consistent dark background across all pages

### Task 6: Update all pages to use dark classes ✅ DONE
- Audit pages: login.vue, index.vue, listings.vue, products.vue, transactions.vue, analytics.vue, scraping.vue, ads.vue
- All pages use dark: component classes

### Task 7: Verify implementation ✅ DONE
- Run dev server and verify dark mode renders correctly
- Check all pages render consistently
- Production build succeeds

## Verification Criteria

- [x] Dark mode is applied by default on page load
- [x] Blue primary colors (primary-500 through primary-900) are visible on dark backgrounds
- [x] All 8 pages render with dark theme
- [x] No unstyled light mode elements remain
- [x] Dev server runs without errors
- [x] Production build succeeds

---

## Spec Status: ✅ DONE

**Completed:** 2026-02-03

**Summary:**
- Dark mode is the default theme with blue accents on dark background
- Tailwind CSS configured with `darkMode: 'class'`
- All component classes use dark-friendly colors (slate-800/900 backgrounds, slate-100 text)
- All 8 pages render correctly with dark theme
- No light mode elements remain
- Production build succeeds
