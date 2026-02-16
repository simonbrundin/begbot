# Dark Mode Shaping Decisions

## Design Decisions

### Color Palette

**Background:**
- Very dark gray/blue instead of pure black for reduced eye strain
- Background: `slate-900` (~#0f172a) or custom hex

**Surface colors:**
- Cards, inputs, modals: `slate-800` (~#1e293b) with lighter borders
- Hover states: `slate-700` (~#334155)

**Text:**
- Primary text: `slate-100` (~#f1f5f9) for readability
- Secondary text: `slate-400` (~#94a3b8)
- Muted text: `slate-500` (~#64748b)

**Primary (Blue Accent) - Keep existing:**
- Primary-500: `#0ea5e9` (main accent)
- Primary-600: `#0284c7` (hover)
- Primary-400: `#38bdf8` (subtle highlights)

### Implementation Approach

1. **CSS Custom Properties vs Tailwind classes**
   - Decision: Use Tailwind's `dark:` variant with class-based dark mode
   - Reason: Already using Tailwind, consistent with existing patterns
   - Apply `dark` class to `<html>` element

2. **Default theme**
   - Decision: Dark mode as default
   - Implementation: Hardcode `dark` class on html element
   - No toggle needed for MVP

3. **Component class strategy**
   - Keep existing component classes (btn, card, input, etc.)
   - Add dark: variants for each
   - Example: `.card` becomes `.dark .card` in CSS

## Context

- **Existing setup:** Tailwind with primary blue colors defined
- **main.css:** Has @layer base with light bg-gray-50
- **Pages:** 8 Vue pages requiring dark mode
- **Components:** No custom components (empty folder)

## Out of Scope

- Light mode support
- Theme toggle/switcher
- System preference detection
- Animations/transitions
