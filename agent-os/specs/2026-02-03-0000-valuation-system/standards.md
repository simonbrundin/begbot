# Standards: Valuation System

## Applicabla standards

### currency-storage
**Path:** `agent-os/standards/database/currency-storage.md`

Monetära värden lagras som heltal (SEK öre).

```go
// ✓ Korrekt
Valuation int  // 1500 = 15.00 kr

// ✗ Fel
Valuation float64  // Aldrig använda float för pengar
```

### swedish-text
**Path:** `agent-os/standards/frontend/swedish-text.md`

All frontend-text ska vara på svenska.

```vue
<!-- ✓ Korrekt -->
<span>Värdering: {{ formatCurrency(valuation) }}</span>

<!-- ✗ Fel -->
<span>Valuation: {{ formatCurrency(valuation) }}</span>
```

### configuration-structure
**Path:** `agent-os/standards/global/configuration-structure.md`

Använd nested config structs med time.Duration för timeouts.

```go
type ValuationConfig struct {
    ExternalServicesTimeout time.Duration `yaml:"external_services_timeout"`
    LLMTimeout              time.Duration `yaml:"llm_timeout"`
    MaxRetries             int           `yaml:"max_retries"`
}
```

### llm-service-functions
**Path:** `agent-os/standards/backend/llm-service-functions.md`

En LLM-funktion per uppgift, Action+Entity naming.

```go
// ✓ Korrekt
func (s *LLMService) CompileValuations(ctx context.Context, valuations []ValuationInput) (*ValuationOutput, error)

// ✗ Fel
func (s *LLMService) GetValuation(productID int64) (Price, Margin, error)
```

## Icke-applicabla standards

- **email-sending:** Inte relevant för denna funktion
