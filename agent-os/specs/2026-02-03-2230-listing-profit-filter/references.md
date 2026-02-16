# References: Listing Profit Calculation and Filtering

## Similar Code Patterns

### Profit Calculation
- `internal/db/postgres.go:457-476` - `CalculateProfit()` for TradedItem
  - Used for sold items calculation
  - Formula: `sellTotal - (buyTotal + sellCost)`

### Valuation Compilation
- `internal/services/valuation.go:193-210` - `CalculateProfit()` and `CalculateProfitMargin()`
  - Similar calculation pattern for buying decisions
  - Used in `ShouldBuy()` function

### Listing Response
- `cmd/api/main.go:160-179` - `getListings()` function
  - Returns `ListingWithValuations` struct
  - Includes listing and valuations data

### Bot Processing
- `internal/services/bot.go:148-282` - `processAd()` function
  - Saves listings after validation
  - Uses `evaluateItem()` for buy decisions

## Database Tables
- `listings` - Stores scraped ads with price, valuation, shipping_cost
- `trading_rules` - Stores min_profit_sek and min_discount

## Models
- `models.Listing` - Core listing model
- `models.Economics` - Contains MinProfitSEK and MinDiscount
