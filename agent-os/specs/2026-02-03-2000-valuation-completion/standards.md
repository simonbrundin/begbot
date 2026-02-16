# Standards: Valuation Completion

## Applicabla standards

### llm-service-functions
**Path:** `agent-os/standards/backend/llm-service-functions.md`

En LLM-funktion per uppgift, Action+Entity naming.

```go
// ✓ Korrekt
func (s *ValuationCompiler) Compile(ctx context.Context, inputs []ValuationInput) (*ValuationOutput, error)

// ✗ Fel
func (s *ValuationService) GetValuation(productID int64) (Price, Margin, error)
```

### configuration-structure
**Path:** `agent-os/standards/global/configuration-structure.md`

Utökad ValuationConfig för nya metoder:

```go
type ValuationConfig struct {
    TargetSellDays          int           `yaml:"target_sell_days"`
    MinProfitMargin         float64       `yaml:"min_profit_margin"`
    CollectionTimeout       time.Duration `yaml:"collection_timeout"`
    MethodTimeouts          map[string]time.Duration `yaml:"method_timeouts"`
    MinConfidenceThreshold  int           `yaml:"min_confidence_threshold"`
    LLMCompilerModel        string        `yaml:"llm_compiler_model"`
}
```

### currency-storage
**Path:** `agent-os/standards/database/currency-storage.md`

Alla monetära värden som heltal (SEK öre).

```go
// ✓ Korrekt
type ValuationInput struct {
    Value int // 1500 = 15.00 kr
}

// ✗ Fel
Value float64 // Aldrig använda float för pengar
```

## Metod-specifika standards

### DatabaseValuationMethod
```
Query-mönster: Använd befintlig GetSoldTradedItems() funktion
Felhantering: Returnera nil resultat, logga fel
Timeout: 5 sekunder
```

### LLMNewPriceMethod
```
Model: Samma modell som för ExtractProductInfo
Timeout: 10 sekunder
Prompt: Svenska, JSON-output
Retry: 1 retry vid timeout
```

### TraderaValuationMethod
```
Source: https://www.tradera.com/valuation
Timeout: 5 sekunder
Error: Returnera nil vid fetch-fel
Logging: Logga URL som försökts
```

### SoldAdsValuationMethod
```
Scraping: Använd befintlig MarketplaceService
Timeout: 8 sekunder
Error: Returnera nil vid fetch-fel
Rate limiting: Respektatera externa servrars gränser
```

## Icke-applicabla standards

- **email-sending:** Inte relevant för valuation
- **swedish-text:** Endast för frontend-visning
