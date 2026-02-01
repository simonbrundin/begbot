package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"begbot/internal/config"
	"begbot/internal/services"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	llmService := services.NewLLMService(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	testAdText := `iPhone 15 Pro Max 256GB. Naturlig titan. Nyskick, köpt för 2 månader sedan. Följer med originallåda och laddare. Skärmen har inga repor, inga bucklor. Säljer för att jag uppgraderar. Pris: 6500 kr.`

	fmt.Println("Testing ExtractProductInfo with z-ai/glm-4.5-air:free model...")
	fmt.Println("Ad text:", testAdText)
	fmt.Println()

	productInfo, err := llmService.ExtractProductInfo(ctx, testAdText, "https://tradera.se/test")
	if err != nil {
		log.Printf("Config default_model: %s", cfg.LLM.DefaultModel)
		log.Printf("Config models: %v", cfg.LLM.Models)
		log.Fatalf("Failed to extract product info: %v", err)
	}

	fmt.Println("Result:")
	fmt.Printf("  Manufacturer: %s\n", productInfo.Manufacturer)
	fmt.Printf("  Model: %s\n", productInfo.Model)
	fmt.Printf("  Category: %s\n", productInfo.Category)
	fmt.Printf("  Storage: %s\n", productInfo.Storage)
	fmt.Printf("  Condition: %s\n", productInfo.Condition)
	fmt.Printf("  NewPrice: %.2f\n", productInfo.NewPrice)
}
