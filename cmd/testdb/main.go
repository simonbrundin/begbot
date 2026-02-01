package main

import (
	"context"
	"fmt"
	"log"

	"begbot/internal/config"
	"begbot/internal/db"
	"begbot/internal/models"
)

func main() {
	// Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	pg, err := db.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pg.Close()

	fmt.Println("✅ Successfully connected to Supabase!")

	// Test: Insert a product
	ctx := context.Background()
	product := models.Product{
		Brand:             "TestBrand",
		Name:              "TestProduct",
		SellPackagingCost: 100, // 1 SEK
		SellPostageCost:   500, // 5 SEK
	}

	err = pg.SaveProduct(ctx, &product)
	if err != nil {
		log.Fatalf("Failed to save product: %v", err)
	}

	fmt.Printf("✅ Saved product with ID: %d\n", product.ID)

	// Test: Query existing product
	retrieved, err := pg.GetProductByName(ctx, "TestBrand", "TestProduct")
	if err != nil {
		log.Fatalf("Failed to retrieve product: %v", err)
	}

	if retrieved != nil {
		fmt.Printf("✅ Retrieved product: Brand=%s, Name=%s\n", retrieved.Brand, retrieved.Name)
	}

	fmt.Println("\n✅ Database is working! You can start saving data.")
	fmt.Println("\n💾 How to use:")
	fmt.Println("  - Import 'begbot/internal/db' and 'begbot/internal/models'")
	fmt.Println("  - Use db.NewPostgres(cfg.Database) to connect")
	fmt.Println("  - Use methods like SaveProduct, SaveTradedItem, etc.")
}
