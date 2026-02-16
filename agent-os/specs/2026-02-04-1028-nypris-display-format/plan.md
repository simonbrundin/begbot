# Plan: Nypris Display Format Change

## Overview
Move "Nypris" display in `/ads` page to be grouped with "Frakt" and "Värdering" instead of being displayed above the listing price. All three should have consistent styling as simple text lines.

## Current State
In `frontend/pages/ads.vue`:
- **Nypris** (lines 103-107): Displayed above listing price with product info
- **Frakt** (lines 121-129): Displayed below listing price
- **Värdering** (lines 130-137): Displayed below Frakt
- **Delvärderingar** (lines 138-149): Displayed in boxes (this is intentional)

## Desired State
- Nypris should appear after listing price, grouped with Frakt and Värdering
- All three should use consistent `text-sm text-slate-400` styling
- Display format: "Label: Value"

## Implementation

### Task 1: Save spec documentation
Create this spec folder with all planning documents.

### Task 2: Move Nypris display location
In `frontend/pages/ads.vue`:
1. Remove Nypris paragraph from lines 103-107 (currently shown above price)
2. Add Nypris display after listing price (around line 120), following same pattern as Frakt/Värdering

### Code Change Example
Remove:
```vue
<p
  v-if="item.Product && item.Product.new_price"
  class="text-sm text-slate-400"
>
  Nypris: {{ formatPriceAsSEK(item.Product.new_price) }}
</p>
```

Add after price display (around line 120):
```vue
<p class="text-sm text-slate-400">
  Nypris: {{ item.Product?.new_price ? formatPriceAsSEK(item.Product.new_price) : "-" }}
</p>
```

## Standards Applied
- **swedish-text**: All UI text in Swedish ("Nypris", "Frakt", "Värdering")
- **components**: Follow existing component patterns in the file
- **css**: Use existing Tailwind utility classes

## Verification
- [ ] Nypris displays below listing price
- [ ] Nypris, Frakt, and Värdering are visually grouped together
- [ ] All three use consistent styling
- [ ] No visual regressions on the ads page
