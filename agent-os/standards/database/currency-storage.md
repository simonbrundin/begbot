# Currency Storage

Store all monetary values as integers representing SEK öre (cents).

**Rule:**
- All price/cost fields use `int` (not `float64`)
- 100 = 1 SEK, 14900 = 149.00 SEK

**When to use float:**
- Display formatting only (divide by 100)
- Never for storage or calculations

**Example:**
```go
type TradedItem struct {
    BuyPrice  int  // 14900 SEK = 149.00 SEK
    SellPrice *int // nullable
}

// Display
fmt.Printf("%.2f SEK", float64(item.BuyPrice)/100)
```
