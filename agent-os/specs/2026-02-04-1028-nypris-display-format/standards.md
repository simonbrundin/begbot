# Standards Applied

## Relevant Standards

### swedish-text
All frontend text must be in Swedish. This applies to:
- "Nypris" label
- "Frakt" label  
- "Värdering" label
- "Okänt" fallback for missing shipping cost

### components
Follow existing Vue component patterns:
- Use existing helper functions (`formatPriceAsSEK`, `formatCurrency`, `formatValuationAsSEK`)
- Use conditional rendering with `v-if` for optional data
- Maintain consistent template structure

### css
Use existing Tailwind CSS utility classes:
- `text-sm` for label/value text size
- `text-slate-400` for secondary text color
- Consistent spacing with existing patterns

## Files to Modify
- `frontend/pages/ads.vue` - Move Nypris display location
