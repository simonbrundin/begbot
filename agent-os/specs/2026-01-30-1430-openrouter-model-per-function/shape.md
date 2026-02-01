# Shaping: OpenRouter Model Selection per Function

## Scope

### Inkluderat
- Endast OpenRouter som provider (ta bort OpenAI-stöd)
- Konfiguration av modell per funktionsnamn
- Default-modell för funktioner utan specifik modell
- OpenRouter-headers (HTTP-Referer, X-Title)
- Statisk config (ingen dynamisk modellhämtning)

### Exkluderat
- OpenAI provider (borttaget)
- Dynamisk modellhämtning vid runtime
- Provider-factory (ingen abstraktion behövs)

## Beslut

### Provider-val
Endast OpenRouter för:
1. Enkel API med OpenAI-kompatibelt format
2. Tillgång till modeller från flera leverantörer
3. Enklare kod utan abstraktionslager

### OpenRouter-headers
```go
headers := map[string]string{
    "HTTP-Referer": siteURL,  // Obligatoriskt för OpenRouter
    "X-Title":     siteName,  // För statistik
}
```

### Konfigurationsstruktur
```yaml
llm:
  provider: "openrouter"
  api_key: "..."
  site_url: "https://..."
  site_name: "Begbot"
  default_model: "openai/gpt-4o"
  models:
    ExtractProductInfo: "anthropic/claude-sonnet-4-20250514"
    ValidateProduct: "google/gemini-2.5-pro"
```

## Kontext

### Befintlig kod att ändra
- `internal/config/config.go`: LLMConfig - lägg till Models-mappning
- `internal/services/llm.go`: Ersätt OpenAI med OpenRouter
- `config.yaml`: Uppdatera med OpenRouter-konfiguration

### Nya filer
- `internal/services/openrouter.go`: OpenRouter-klient
