# Shaping: Valuation Completion

## Scope

### In Scope
- Färdigställa alla värderingsmetoder i `ValuationService`
- Skapa plugin-arkitektur för enkelt tillägg av nya metoder
- LLM-kompilering till pris + säkerhetsprocent
- Integration i `BotService.processAd()` arbetsflöde
- Spara alla individuella värderingar till databas

### Out of Scope
- Realtime scraping av externa källor (timeout efter 5s)
- Historik-grafer i UI
- Automatisk återförsök av misslyckade metoder
- Caching av externa värderingar (version 2)

## Designbeslut

### 1. ValuationMethod Interface

```go
type ValuationMethod interface {
    // Unikt namn för metoden
    Name() string

    // Kör värdering och returnera resultat
    // ctx: cancellation context
    // productID: produkten som ska värderas (0 om ny produkt)
    // productInfo: textbeskrivning från LLM
    Valuate(ctx context.Context, productID int64, productInfo string) (*ValuationInput, error)

    // Metodens prioritet (lägre = högre prioritet)
    Priority() int
}

type ValuationInput struct {
    Method       string         // Namn från Name()
    Value        int            // Pris i öre
    Currency     string         // "SEK"
    Confidence   int            // 0-100 procent
    SourceURL    string         // Källa (valfritt)
    Metadata     map[string]any // Extra data
    CollectedAt  time.Time
}
```

### 2. ValuationService med Registry

```go
type ValuationService struct {
    cfg           *config.Config
    database      *db.Postgres
    llmService    *LLMService
    methods       []ValuationMethod
    compiler      *ValuationCompiler
}

func (s *ValuationService) RegisterMethod(m ValuationMethod) {
    s.methods = append(s.methods, m)
    sort.Slice(s.methods, func(i, j int) bool {
        return s.methods[i].Priority() < s.methods[j].Priority()
    })
}

func (s *ValuationService) CollectAll(ctx context.Context, productID int64, productInfo string) ([]ValuationInput, error) {
    var results []ValuationInput
    for _, method := range s.methods {
        if ctx.Err() != nil {
            break
        }
        result, err := method.Valuate(ctx, productID, productInfo)
        if err != nil {
            log.Printf("Valuation method %s failed: %v", method.Name(), err)
            continue
        }
        results = append(results, *result)
    }
    return results, nil
}
```

### 3. DatabaseValuationMethod (linjär regression)

```
Steg:
1. Hämta sålda items från traded_items (senaste N st)
2. Extrahera (days_on_market, sell_price) för varje
3. Beräkna linjär regression: price = intercept + k * days
4. Prediktera pris för target_days (från config)
5. Confidence: baserat på R² + antal datapunkter
   - <10 items: 30%
   - 10-50 items: 60%
   - >50 items: 85%
```

### 4. LLMNewPriceMethod

```
Prompt till LLM:
"""
Given this product:
{product_info}

Estimate the NEW retail price in Swedish kronor (SEK).
Consider:
- Current market conditions
- Brand and model
- Product age and condition

Return ONLY a JSON object:
{"price": 1500, "confidence": 75, "reasoning": "..."}
"""

Confidence-faktorer:
- LLM:s egen confidence (extraherad från svar)
- Produktinfo-kvalitet
- Tillgänglig produktdata
```

### 5. ValuationCompiler (LLM-kompilering)

```
Input:
[
  {method: "Database", value: 1350, confidence: 70, n: 45},
  {method: "Tradera", value: 1420, confidence: 50},
  {method: "LLM New Price", value: 1600, confidence: 60},
  {method: "eBay Sold", value: 1380, confidence: 65}
]

Prompt:
"""
These are valuations for {product}:

1. Database analysis (45 sold items): 1350 SEK (confidence: 70%)
2. Tradera valuation: 1420 SEK (confidence: 50%)
3. New price (LLM): 1600 SEK (confidence: 60%)
4. eBay sold listings: 1380 SEK (confidence: 65%)

Suggest:
1. Recommended selling price (integer SEK)
2. Overall confidence percentage (0-100)
3. Brief reasoning for the decision

Return ONLY JSON:
{"price": 1400, "confidence": 65, "reasoning": "..."}
"""
```

### 6. Integration i processAd()

```
processAd() workflow:
1. Fetch ad details ✓
2. Extract product info via LLM ✓
3. [NYTT] CollectValuations(productID, productInfo)
   - Kör alla registrerade metoder
   - Spara varje ValuationInput till databas
   - Returnera []ValuationInput
4. [NYTT] CompileValuations([]ValuationInput)
   - LLM-kompilering
   - Returnera: Price, Confidence, Reasoning
5. Evaluate item med kompilerat pris
6. Save listing ✓
```

## Öppna frågor

1. **Hur hantera timeout per metod?**
   - Sätt global timeout (10s) + per-metod timeout via context
   - Timeout-känsliga metoder (database) får korta timeouts

2. **Hur hantera motstridiga värderingar?**
   - LLM compiler väger samman med confidence
   - Manuell granskning om confidence < 30%

3. **Ska vi cacha externa värderingar?**
   - Ja, i `valuations` tabellen med expire-tid
   - Redis cache för < 1h gamla värderingar (version 2)

## Relaterade filer att modifiera

- `internal/services/valuation.go` - Ny arkitektur
- `internal/services/bot.go:147` - processAd() integration
- `internal/db/postgres.go` - Nya queries om nödvändigt
- `cmd/api/main.go` - Nya endpoints
- `internal/config/config.go` - Valuation config utökning
