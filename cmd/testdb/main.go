package main

import (
	"context"
	"fmt"
	"log"

	"begbot/internal/config"
	"begbot/internal/db"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pg, err := db.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pg.Close()

	ctx := context.Background()
	listings, err := pg.GetAllListings(ctx)
	if err != nil {
		log.Fatalf("GetAllListings failed: %v", err)
	}

	fmt.Printf("Found %d listings\n", len(listings))
	for i, l := range listings {
		if i > 5 {
			break
		}
		title := l.Title
		desc := ""
		if l.Description != nil {
			desc = *l.Description
		}
		fmt.Printf("Listing %d: title='%s' desc_len=%d\n", l.ID, title[:min(50, len(title))], len(desc))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
