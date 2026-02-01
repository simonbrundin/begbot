# Standards: Search Terms Feature

## Applicable Standards

### currency-storage
All price fields must be stored as integers representing SEK öre.
- `max_price` field in search_criteria: integer (e.g., 15000 = 150.00 SEK)
- Use `*int` for nullable values

Reference: `agent-os/standards/database/currency-storage.md`

### configuration-structure
Configuration for search terms should follow nested struct patterns.
- Group marketplace-specific settings
- Use time.Duration for timeout settings

Reference: `agent-os/standards/global/configuration-structure.md`

## Not Applicable

- email-sending: Not required for this feature
- llm-service-functions: Not required for this feature

## New Standards to Consider

### Search Term Naming
- Use descriptive names for search terms
- Include product category in name for clarity
- Example: "iPhone 13-15 Stockholm", "Lego Star Wars"

### URL Building
- Validate URLs before saving
- Log warnings for malformed URLs
- Use URL struct for parsing and manipulation

### Marketplace-Specific Parameters
Store marketplace-specific parameters as JSON in `extra_params` column.

**Common fields (standardized):**
- `max_price`: Maximum price in SEK öre
- `min_condition`: Minimum condition ID

**Blocket example extra_params:**
```json
{"shipping_types": "0", "sort": "PUBLISHED_DESC", "product_category": "2.93.3217.39"}
```

**Tradera example extra_params:**
```json
{"categoryId": 123, "conditionId": 2, "sortOrder": "ending"}
```

**URL building flow:**
1. Get base URL from marketplaces table
2. Add common params (q, price_from, condition)
3. Merge extra_params JSON as URL query parameters
