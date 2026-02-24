package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"begbot/internal/config"
	"begbot/internal/db"
	"begbot/internal/models"
)

type BlocketAdResponse struct {
	LoaderData struct {
		ItemRecommerce struct {
			ItemData struct {
				Category CategoryNode `json:"category"`
			} `json:"itemData"`
		} `json:"item-recommerce"`
	} `json:"loaderData"`
}

type CategoryNode struct {
	ID     int64         `json:"id"`
	Value  string        `json:"value"`
	Parent *CategoryNode `json:"parent"`
}

type CategoryInfo struct {
	Root *CategoryNode
	Leaf *CategoryNode
}

func main() {
	ctx := context.Background()

	database, err := db.NewPostgres(config.DatabaseConfig{
		Host:     getEnv("DATABASE_HOST", "aws-1-eu-west-1.pooler.supabase.com"),
		Port:     5432,
		User:     getEnv("DATABASE_USER", "postgres.fxhknzpqhrkpqothjvrx"),
		Password: getEnv("DATABASE_PASSWORD", ""),
		Name:     getEnv("DATABASE_NAME", "postgres"),
		SSLMode:  getEnv("DATABASE_SSLMODE", "require"),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	products, err := getAllProducts(ctx, database)
	if err != nil {
		log.Fatalf("Failed to get products: %v", err)
	}

	log.Printf("Found %d products", len(products))

	httpClient := &http.Client{Timeout: 10 * time.Second}

	for _, product := range products {
		listing, err := getListingForProduct(ctx, database, product.ID)
		if err != nil {
			log.Printf("Failed to get listing for product %d: %v", product.ID, err)
			continue
		}
		if listing == nil {
			log.Printf("No listing found for product %d (%s %s)", product.ID, *product.Brand, *product.Name)
			continue
		}

		adID := extractAdID(listing.Link)
		if adID == "" {
			log.Printf("No ad ID found in link for product %d: %s", product.ID, listing.Link)
			continue
		}

		log.Printf("Product %d (%s %s): checking Blocket ad %s", product.ID, *product.Brand, *product.Name, adID)

		catInfo, err := fetchBlocketCategory(httpClient, adID)
		if err != nil {
			log.Printf("Failed to fetch category for ad %s: %v", adID, err)
			continue
		}
		if catInfo == nil {
			log.Printf("No category info for ad %s", adID)
			continue
		}

		// Build category ID dynamically by traversing from leaf to root
		// Format: 2.{...parent_ids}.{leaf_id}
		// Example: 2.93.3217.39 (Elektronik -> Telefoner -> Mobiltelefoner)
		blocketID := buildBlocketCategoryID(catInfo)

		log.Printf("  Category: %s - blocket_id: %s", getCategoryPath(catInfo), blocketID)

		llmCategory := ""
		if product.Category != nil {
			llmCategory = *product.Category
		}

		err = saveBlocketCategory(ctx, database, blocketID, getCategoryName(catInfo), llmCategory)
		if err != nil {
			log.Printf("Failed to save category %s: %v", blocketID, err)
			continue
		}

		// Update the product with the correct blocket_category
		_, err = database.DB().ExecContext(ctx, "UPDATE products SET blocket_category = $1 WHERE id = $2", blocketID, product.ID)
		if err != nil {
			log.Printf("Failed to update product %d with blocket_category %s: %v", product.ID, blocketID, err)
			continue
		}

		log.Printf("  Saved: blocket_id=%s, name=%s, llm_category=%s, updated product %d", blocketID, getCategoryName(catInfo), llmCategory, product.ID)

		time.Sleep(200 * time.Millisecond)
	}

	log.Println("Done!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getAllProducts(ctx context.Context, database *db.Postgres) ([]models.Product, error) {
	query := `SELECT id, brand, name, category, model_variant, sell_packaging_cost, sell_postage_cost, new_price, enabled, created_at FROM products`
	rows, err := database.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.Brand, &p.Name, &p.Category, &p.ModelVariant, &p.SellPackagingCost, &p.SellPostageCost, &p.NewPrice, &p.Enabled, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func getListingForProduct(ctx context.Context, database *db.Postgres, productID int64) (*models.Listing, error) {
	query := `SELECT id, product_id, price, link, title, description, marketplace_id, status, publication_date, is_my_listing, shipping_cost, shipping_insurance 
		FROM listings WHERE product_id = $1 AND status = 'active' LIMIT 1`
	var listing models.Listing
	err := database.DB().QueryRowContext(ctx, query, productID).Scan(
		&listing.ID, &listing.ProductID, &listing.Price, &listing.Link, &listing.Title, &listing.Description,
		&listing.MarketplaceID, &listing.Status, &listing.PublicationDate, &listing.IsMyListing,
		&listing.ShippingCost, &listing.ShippingInsurance,
	)
	if err != nil {
		return nil, err
	}
	return &listing, nil
}

func extractAdID(link string) string {
	re := regexp.MustCompile(`/item/(\d+)`)
	matches := re.FindStringSubmatch(link)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func fetchBlocketCategory(client *http.Client, adID string) (*CategoryInfo, error) {
	url := fmt.Sprintf("https://blocket-api.se/v1/ad/recommerce?id=%s", adID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result BlocketAdResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	cat := result.LoaderData.ItemRecommerce.ItemData.Category
	if cat.ID == 0 {
		return nil, nil
	}

	// The API returns the category with leaf at top and parents nested
	// We need to traverse from the leaf (top) up to the root to build the full path
	info := &CategoryInfo{
		Leaf: &cat, // The API response IS the leaf with parents attached
	}

	log.Printf("    Raw category: %s", getCategoryPath(info))

	return info, nil
}

func saveBlocketCategory(ctx context.Context, database *db.Postgres, blocketID, name, llmCategory string) error {
	query := `
		INSERT INTO blocket_categories (blocket_id, name, llm_category)
		VALUES ($1, $2, $3)
		ON CONFLICT (blocket_id) DO UPDATE SET
			name = EXCLUDED.name,
			llm_category = COALESCE(blocket_categories.llm_category, EXCLUDED.llm_category)
	`
	_, err := database.DB().ExecContext(ctx, query, blocketID, name, llmCategory)
	return err
}

func buildBlocketCategoryID(info *CategoryInfo) string {
	if info == nil || info.Leaf == nil {
		return ""
	}

	// Collect all IDs from leaf to root (API gives us leaf with parents attached)
	var ids []int64
	current := info.Leaf
	for current != nil {
		if current.ID > 0 {
			ids = append([]int64{current.ID}, ids...)
		}
		current = current.Parent
	}

	// Build the category ID string: 2.{...all_ids...}
	parts := make([]string, len(ids)+1)
	parts[0] = "2"
	for i, id := range ids {
		parts[i+1] = fmt.Sprintf("%d", id)
	}

	return strings.Join(parts, ".")
}

func getCategoryPath(info *CategoryInfo) string {
	if info == nil || info.Leaf == nil {
		return ""
	}

	var path []string
	current := info.Leaf
	for current != nil {
		if current.ID > 0 {
			path = append([]string{current.Value}, path...)
		}
		current = current.Parent
	}

	return strings.Join(path, " -> ")
}

func getCategoryName(info *CategoryInfo) string {
	if info == nil || info.Leaf == nil {
		return ""
	}
	// Return the leaf category name (most specific)
	return info.Leaf.Value
}
