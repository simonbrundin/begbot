# Standards: Listing Profit Calculation and Filtering

## Applicable Standards

### database/currency-storage.md
- Store all monetary values as integers (SEK öre)
- `price`, `valuation`, `shipping_cost`, `profit` are all integers

### global/validation.md
- Validate profit calculations don't produce negative values incorrectly
- Handle nil valuations gracefully

### backend/api.md
- Return profit/discount in listing response
- Consistent field naming across API

## Validation Rules

1. **Profit calculation**:
   - Skip if `valuation` is nil
   - Skip if `price` is nil
   - Shipping cost defaults to 0 if nil

2. **Discount calculation**:
   - Division by zero protection
   - Return 0 if valuation is 0

3. **Filtering**:
   - Use values from `trading_rules` table
   - Default to 0 if rules not configured
