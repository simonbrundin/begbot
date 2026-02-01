# LLM Service Functions

**Rule:**
- En funktion per LLM-uppgift (ExtractProductInfo, ValidateProduct, etc.)
- Använd Action+Entity naming (ExtractProductInfo, ValidateProduct)
- Varje funktion har specifik prompt
- Returnera strukturerad data (inte råa strängar)

**Varför:**
- Bättre prompts för varje specifik uppgift
- Enklare att felsöka och testa

**Exempel:**
```go
type LLMService struct {
    cfg *config.Config
}

type ProductInfo struct {
    Manufacturer string
    Model        string
    Storage      string
    Condition    string
    ShippingCost float64
    NewPrice     float64
}

func (s *LLMService) ExtractProductInfo(ctx context.Context, adText, link string) (*ProductInfo, error) {
    // specifik prompt för produktextraktion
}

func (s *LLMService) ValidateProduct(ctx context.Context, productInfo *ProductInfo, adText string) (bool, string, error) {
    // specifik prompt för validering
}
```
