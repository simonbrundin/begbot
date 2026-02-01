# References: Listing Product Validation

## Kodreferenser

### Databaslager

| Fil | Rad | Beskrivning |
|-----|-----|-------------|
| `internal/db/postgres.go` | 292-299 | `SaveProduct()` - Spara produkt |
| `internal/db/postgres.go` | 301-317 | `GetProductByName()` - Hämta produkt efter brand+name |
| `internal/db/postgres.go` | 319-338 | `GetOrCreateProduct()` - Pattern för find-or-create |
| `internal/db/postgres.go` | 244-254 | `SaveListing()` - Spara listing |

### Tjänstelager

| Fil | Rad | Beskrivning |
|-----|-----|-------------|
| `internal/services/bot.go` | 89-96 | `processAd()` - Huvudflöde för annonshantering |
| `internal/services/bot.go` | 92-96 | `ExtractProductInfo()` - Produktinfo-extraktion |
| `internal/services/llm.go` | 31-33 | `ValidateProduct()` - Befintlig LLM-validering (stubbad) |

### Models

| Fil | Fält | Beskrivning |
|-----|------|-------------|
| `internal/models/models.go` | `Product` | Befintlig modell (behöver utökas) |

## Nya Funktioner att Skapa

### Databas
```go
func (p *Postgres) FindProduct(ctx context.Context, brand, name, category string) (*models.Product, error)
```

### Tjänstelager
```go
func (s *LLMService) ExtractProductInfo(...) (*ProductInfo, error)  // Utök med Category
func (s *BotService) ValidateListing(...) error                      // Validering
```

## Likheter med Befintlig Kod

Denna feature följer samma mönster som `ExtractProductInfo()` + `ValidateProduct()`:
1. Anropa LLM för att extrahera produktinfo
2. Validera mot databas/affärsregel
3. Returnera error/continue om invalid

## Mönster att Följa

```go
// Befintligt mönster i bot.go:92-96
productInfo, err := s.llmService.ExtractProductInfo(ctx, ad.AdText, ad.Link)
if err != nil {
    log.Printf("Failed to extract product info: %v", err)
    return err
}
```
