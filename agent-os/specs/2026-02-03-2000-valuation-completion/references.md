# References: Valuation Completion

## Intern kod att referera till

### ValuationService (befintlig)
**Fil:** `internal/services/valuation.go`
**Användning:** Bas att bygga vidare på
**Viktiga delar:**
- `GetHistoricalValuation()` - Linjär regression mönster (rad 23)
- `CollectDatabaseValuation()` - Befintlig implementation (rad 126)
- `CollectLLMNewPrice()` - Stub att ersätta (rad 170)
- `ValuationInput`, `ValuationOutput` structs (rad 111-124)

### BotService.processAd()
**Fil:** `internal/services/bot.go:147`
**Användning:** Integration-point för valuation
**Workflow:**
- `ExtractProductInfo()` (rad 160) - hämta produktinfo
- `evaluateItem()` (rad 186) - använd valuation
- Sätt in `CollectValuations()` mellan dessa

### Database-lager
**Fil:** `internal/db/postgres.go`
**Tabeller:**
- `valuations` (rad 169) - spara individuella värderingar
- `valuation_types` (rad 165) - typ-referens
**Funktioner:**
- `GetSoldTradedItems()` (rad ?) - hämta sålda items
- `CreateValuation()` (rad ?) - spara värdering

### Config
**Fil:** `internal/config/config.go`
**Struktur:**
```go
type Config struct {
    // ... befintliga fält
    Valuation ValuationConfig
}
```

### API handlers
**Fil:** `cmd/api/main.go`
**Mönster:**
- `valuationsHandler()` (rad 548) - befintlig handler
- Lägg till `POST /api/valuations/collect`

## Externa resurser

### Tradera Valuation
**URL:** https://www.tradera.com/valuation
**Användning:** Referens för prisdata-format

### Linjär regression
**Fil:** `internal/services/valuation.go:54-64`
**Implementerad:** `kValue` och `intercept` beräkning
**Användning:** DatabaseValuationMethod bas

## Mönster att följa

1. **Felhantering i services:**
   ```go
   if err != nil {
       log.Printf("Method %s failed: %v", m.Name(), err)
       return nil, nil
   }
   ```

2. **Context-timeout:**
   ```go
   select {
   case <-ctx.Done():
       return nil, ctx.Err()
   case <-time.After(methodTimeout):
       return nil, fmt.Errorf("timeout")
   }
   ```

3. **Sorting av metoder:**
   ```go
   sort.Slice(methods, func(i, j int) bool {
       return methods[i].Priority() < methods[j].Priority()
   })
   ```

## Kod att studera

1. `services/bot.go:247-267` - `evaluateItem()` som använder valuation
2. `services/valuation.go:23-72` - Linjär regression implementering
3. `services/marketplace.go` - Scraping-mönster för SoldAds
