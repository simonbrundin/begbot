# References

## Reference Implementation

### Similar Code in Codebase
The pattern for displaying price information is already established in the same file:

**Frakt display (lines 121-129):**
```vue
<p class="text-sm text-slate-400">
  Frakt:
  {{
    item.Listing.shipping_cost !== null &&
    item.Listing.shipping_cost !== undefined
      ? formatCurrency(item.Listing.shipping_cost)
      : "Okänt"
  }}
</p>
```

**Värdering display (lines 130-137):**
```vue
<p class="text-sm text-slate-400">
  Värdering:
  {{
    item.Listing.valuation
      ? formatValuationAsSEK(item.Listing.valuation)
      : "-"
  }}
</p>
```

**Nypris current display (lines 103-107):**
```vue
<p
  v-if="item.Product && item.Product.new_price"
  class="text-sm text-slate-400"
>
  Nypris: {{ formatPriceAsSEK(item.Product.new_price) }}
</p>
```

## Helper Functions
- `formatPriceAsSEK(price: number | null)` - Formats price with Swedish locale
- `formatCurrency(price: number | null)` - Alternative formatting function
- `formatValuationAsSEK(sek: number | null)` - Valuation-specific formatting
