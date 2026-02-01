# Standards: OpenRouter Model Selection per Function

## Tillämpliga standards

### configuration-structure
**Applicerad:** Nested structs för LLM-konfiguration
```go
type LLMConfig struct {
    Provider       string            `yaml:"provider"`
    DefaultModel   string            `yaml:"default_model"`
    Models         map[string]string `yaml:"models"` // funktion -> modell
}
```
`time.Duration` används för `timeout`.

### llm-service-functions
**Fortsatt tillämpning:** En funktion per LLM-uppgift
- `ExtractProductInfo()` - produktextraktion
- `ValidateProduct()` - validering
- `CheckProductCondition()` - konditionskontroll
- `EstimateNewPrice()` - prisuppskattning

Varje funktion behåller sin specifika prompt men får nu dynamisk modell.

## Ej applicerade standards

### email-sending
Inte relevant för denna feature.

### currency-storage
Inte relevant för denna feature.

## Nya konventioner

### Model ID-format
Använd OpenRouter-format: `provider/model-id`
Exempel: `openai/gpt-4o`, `anthropic/claude-sonnet-4-20250514`

### Config-uppdatering
För att uppdatera befintlig config:
```yaml
# gammal:
llm:
  model: "gpt-4"

# ny:
llm:
  provider: "openrouter"
  default_model: "openai/gpt-4o"
  models:
    ExtractProductInfo: "anthropic/claude-sonnet-4-20250514"
```
