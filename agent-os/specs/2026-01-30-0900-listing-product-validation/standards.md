# Standards: Listing Product Validation

## Tillämpade Standards

### backend/llm-service-functions.md

**Relevans**: Hög

Valideringen använder LLM för att extrahera produktinfo inklusive kategori, vilket följer standarden "En LLM-funktion per uppgift".

```go
// Följer mönstret: Action+Entity naming
productInfo, err := s.llmService.ExtractProductInfo(ctx, ad.AdText, ad.Link)
```

**Category-förslag** att LLM ska returnera:
- `phone` - Smartphones
- `tablet` - Tablets
- `watch` - Smartwatches
- `headphones` - Headphones/earbuds
- `case` - Phone cases/covers
- `charger` - Chargers/cables
- `accessory` - Other accessories
- `computer` - Laptops/desktops
- `component` - Spare parts

### database/currency-storage.md

**Relevans**: Låg (ingen valutahantering i denna feature)

## Ej Tillämpade Standards

Inga andra standards är direkt applicerbara för denna feature.
