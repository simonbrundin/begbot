package main

import (
	"log"
	"os"
	"time"

	"begbot/internal/config"
	"begbot/internal/db"
	"begbot/internal/services"

	"github.com/joho/godotenv"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func main() {
	godotenv.Load()
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var database *db.Postgres
	for i := 0; i < 3; i++ {
		database, err = db.NewPostgres(cfg.Database)
		if err == nil {
			break
		}
		log.Printf("Database connection attempt %d failed: %v, retrying...", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database after retries: %v", err)
	}
	defer database.Close()

	log.Println("Running database migrations...")
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully!")

	marketplaceService := services.NewMarketplaceService(cfg)
	cacheService := services.NewCacheService(cfg)
	llmService := services.NewLLMService(cfg)
	valuationService := services.NewValuationService(cfg, database, llmService)
	botService := services.NewBotService(cfg, marketplaceService, cacheService, llmService, valuationService, database)

	log.Println("Starting ad fetch...")
	if err := botService.Run(); err != nil {
		log.Fatalf("Ad fetch failed: %v", err)
	}

	log.Println("Ad fetch completed successfully!")
}
