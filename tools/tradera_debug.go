package main

import (
	"fmt"
	"time"

	"begbot/internal/config"
	"begbot/internal/services"
)

func main() {
	cfg := &config.Config{}
	cfg.Scraping.Tradera.Enabled = true
	cfg.Scraping.Tradera.Timeout = 20 * time.Second
	// leave BaseURL empty to use production Tradera

	pi := services.ProductInfo{Manufacturer: "Apple", Model: "iPad mini"}
	fmt.Println("Running Tradera valuation for:", pi)
	v, err := services.RunTraderaValuation(cfg, pi)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Result: %#v\n", v)
}
