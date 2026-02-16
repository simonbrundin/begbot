# Plan: Listing Profit Calculation and Filtering

## Overview
Add profit calculation to listings display and filter scraped listings based on trading rules.

## Requirements
1. **Show profit potential**: Calculate and display profit for each listing
   - Formula: `profit = valuation - price - shipping_cost`
   - Discount: `discount = profit / valuation` (percentage)

2. **Filter during scraping**: Only save listings that meet trading rules criteria:
   - `profit > min_profit_sek`
   - `discount >= min_discount`

## Files to Modify

### Database
- `internal/db/postgres.go`:
  - Add `GetTradingRules()` function

### Models
- `internal/models/models.go`:
  - Add `PotentialProfit` and `Discount` fields to `Listing` struct or create new response struct

### API
- `cmd/api/main.go`:
  - Calculate profit/discount when returning listings

### Bot Service
- `internal/services/bot.go`:
  - Add `GetTradingRules()` call
  - Filter listings before saving based on profit/discount thresholds

## Implementation Order

### Task 1: Save spec documentation (DONE)

### Task 2: Add GetTradingRules() DB function
**File**: `internal/db/postgres.go`
- Add `GetTradingRules(ctx context.Context) (*models.Economics, error)`
- Fetch from `trading_rules` table
- Return MinProfitSEK and MinDiscount

### Task 3: Create ListingWithProfit struct
**File**: `internal/db/postgres.go`
- Extend `ListingWithValuations` with profit fields
- Add `PotentialProfit` (int)
- Add `DiscountPercent` (float64)

### Task 4: Calculate profit in GetListingsWithValuations
**File**: `internal/db/postgres.go`
- Update `GetListingsWithValuations()` to calculate profit
- Formula: `valuation - price - shipping_cost`
- Formula: `(profit / valuation) * 100` for discount

### Task 5: Update API response
**File**: `cmd/api/main.go`
- Update `getListings()` to include profit fields in response
- Ensure profit is calculated before JSON encoding

### Task 6: Add trading rules to BotService
**File**: `internal/services/bot.go`
- Add `tradingRules` field to `BotService`
- Fetch rules in `Run()` or `processAd()`
- Pass rules to `processAd()`

### Task 7: Filter listings in processAd
**File**: `internal/services/bot.go`
- Before calling `SaveListing()`:
  - Check if `listing.Valuation` is set
  - Calculate profit: `valuation - price - shipping_cost`
  - Calculate discount: `(profit / valuation) * 100`
  - Skip if `profit <= min_profit_sek` OR `discount < min_discount`
- Log skipped listings with reason

### Task 8: Test the implementation
- Run `go build` to verify compilation
- Run existing tests
- Manual test with API endpoint
