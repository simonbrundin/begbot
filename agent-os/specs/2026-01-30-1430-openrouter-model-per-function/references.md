# References: OpenRouter Model Selection per Function

## Kodreferenser

### Konfigurationslager

| Fil | Rad | Beskrivning |
|-----|-----|-------------|
| `internal/config/config.go` | 48-53 | `LLMConfig` - Befintlig struktur |
| `config.yaml` | 21-25 | `llm:` - Befintlig LLM-konfiguration |

### Tjänstelager

| Fil | Rad | Beskrivning |
|-----|-----|-------------|
| `internal/services/llm.go` | 15-23 | `LLMService` - Huvudstruktur |
| `internal/services/llm.go` | 36-69 | `ExtractProductInfo()` - Referens för funktion |
| `internal/services/llm.go` | 71-73 | `ValidateProduct()` - Referens för funktion |
| `internal/services/llm.go` | 75-77 | `CheckProductCondition()` - Referens för funktion |
| `internal/services/llm.go` | 79-81 | `EstimateNewPrice()` - Referens för funktion |

## Externa resurser

### OpenRouter API
- Base URL: `https://openrouter.ai/api/v1`
- Format: OpenAI-kompatibelt
- Docs: https://openrouter.ai/docs

### Exempel OpenRouter API-anrop
```bash
curl https://openrouter.ai/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENROUTER_API_KEY" \
  -H "HTTP-Referer: $YOUR_SITE_URL" \
  -H "X-Title: $YOUR_SITE_NAME" \
  -d '{
    "model": "openai/gpt-4o",
    "messages": [
      {"role": "user", "content": "What is the capital of Sweden?"}
    ]
  }'
```

## OpenRouter-modeller att överväga

| Kategori | Modell | ID | Användning |
|----------|--------|-----|-----------|
| Standard | GPT-4o | `openai/gpt-4o` | Allround |
| Snabb | GPT-4o-mini | `openai/gpt-4o-mini` | Enkla uppgifter |
| Reasoning | Claude Sonnet 4 | `anthropic/claude-sonnet-4-20250514` | Komplex analys |
| Ekonomisk | DeepSeek Chat | `deepseek/deepseek-chat` | Kostnadskänslig |
| Vision | Gemini 2.5 Pro | `google/gemini-2.5-pro` | Bildanalys |

## Nya filer att skapa

### `internal/services/openrouter.go`
```go
type OpenRouterClient struct {
    apiKey string
    baseURL string
}

func NewOpenRouterClient(apiKey string) *OpenRouterClient
func (c *OpenRouterClient) Chat(ctx context.Context, model, prompt string) (string, error)
```

### `internal/services/openrouter_models.go`
```go
func GetAvailableModels() []string // Hämta från OpenRouter API
```

## Mönster att följa

Befintligt mönster i `llm.go:48-55`:
```go
resp, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
    Messages: []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage(prompt),
    },
    Model:       openai.ChatModel(s.cfg.LLM.Model),
    MaxTokens:   param.NewOpt(int64(200)),
    Temperature: param.NewOpt(0.1),
})
```

Samma format fungerar med OpenRouter (OpenAI-kompatibelt).
