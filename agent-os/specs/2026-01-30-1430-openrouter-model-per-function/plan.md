# Plan: OpenRouter Model Selection per Function

## Översikt
Ersätt OpenAI med OpenRouter som enda LLM-provider. Lägg till stöd för att välja olika modeller per funktion via konfiguration.

## Task 1: Save spec documentation
- [x] Spara denna plan
- [x] Spara shape.md
- [x] Spara standards.md
- [x] Spara references.md

## Task 2: Uppdatera LLMConfig struktur
Modifiera `internal/config/config.go`:
```go
type LLMConfig struct {
    Provider     string            `yaml:"provider"`
    APIKey       string            `yaml:"api_key"`
    SiteURL      string            `yaml:"site_url"`      // för OpenRouter HTTP-Referer
    SiteName     string            `yaml:"site_name"`     // för OpenRouter X-Title
    Timeout      time.Duration     `yaml:"timeout"`
    DefaultModel string            `yaml:"default_model"`
    Models       map[string]string `yaml:"models"` // funktion -> modell
}
```

Uppdatera `config.yaml`:
```yaml
llm:
  provider: "openrouter"
  api_key: "your_openrouter_key"
  site_url: "https://din-domän.se"
  site_name: "Begbot"
  default_model: "openai/gpt-4o"
  timeout: 60s
  models:
    ExtractProductInfo: "anthropic/claude-sonnet-4-20250514"
    ValidateProduct: "google/gemini-2.5-pro"
    CheckProductCondition: "openai/gpt-4o-mini"
    EstimateNewPrice: "deepseek/deepseek-chat"
```

**Status: ✅ COMPLETED**

## Task 3: Skapa OpenRouter-klient
Skapa `internal/services/openrouter.go`:
```go
type OpenRouterClient struct {
    apiKey   string
    siteURL  string
    siteName string
}

func NewOpenRouterClient(apiKey, siteURL, siteName string) *OpenRouterClient
func (c *OpenRouterClient) Chat(ctx context.Context, model, prompt string) (string, error)
```

**Status: ✅ COMPLETED**

## Task 4: Refaktorera LLMService
Modifiera `internal/services/llm.go`:
- Ta bort OpenAI direct import
- Använd OpenRouter-klient istället
- Varje funktion läser modell från `cfg.LLM.Models[funktionsnamn]` eller `cfg.LLM.DefaultModel`

**Status: ✅ COMPLETED**

## Task 5: Uppdatera config.yaml
Lägg till OpenRouter-specifika fält och models-mappning.

**Status: ✅ COMPLETED**

## Task 6: Skriv tester
- Testa config-loading med modell-mappning
- Testa OpenRouter API-anrop (mockat)

**Status: ✅ COMPLETED**
- Tests in `internal/services/openrouter_test.go`
- Tests in `internal/config/config_test.go`

## Task 7: Verifiera och lint
- Kör `go vet`
- Kör `gofmt`

**Status: ✅ COMPLETED**

## Definition of Done

- [x] LLMConfig has Models map for per-function model selection
- [x] OpenRouter client implemented with proper headers
- [x] LLMService uses OpenRouter for all LLM calls
- [x] config.yaml configured with OpenRouter and model mappings
- [x] Unit tests pass for OpenRouter client and config loading
- [x] `go vet` passes
- [x] `gofmt` passes
