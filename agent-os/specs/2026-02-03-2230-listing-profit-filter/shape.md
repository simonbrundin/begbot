# Shape: Listing Profit Calculation and Filtering

## Scope

### What's In
- Calculate profit and discount for each listing
- Filter scraped listings based on trading rules thresholds
- Display profit/discount in API response

### What's Out
- No changes to database schema (use existing fields)
- No changes to frontend (API only)
- No automatic buying functionality

## Key Decisions

### Profit Calculation Location
**Decision**: Calculate profit in API layer, not database
- Reason: Valuation might not be present when saving listing
- Pro: Flexible, can recalculate on demand
- Con: Slight overhead on each request

### Trading Rules Storage
**Decision**: Use existing `trading_rules` table
- Table already exists with `min_profit_sek` and `min_discount`
- Need to add DB function to fetch rules

### Filtering Strategy
**Decision**: Filter before saving in `processAd()`
- Check profit and discount before calling `SaveListing()`
- Log when listings are skipped
- Only save listings that meet both criteria

## Formula

```
profit = valuation - price - shipping_cost
discount_percent = (profit / valuation) * 100
```

## Thresholds
From `trading_rules` table:
- `min_profit_sek`: Minimum profit required (integer, SEK)
- `min_discount`: Minimum discount percentage (integer, 0-100)

## Save Criteria
```
profit > min_profit_sek AND discount_percent >= min_discount
```

## Response Format
```json
{
  "listing": { ... },
  "potential_profit": 500,
  "discount_percent": 25
}
```
