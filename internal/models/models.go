package models

import (
	"encoding/json"
	"time"
)

type Product struct {
	ID                int64      `json:"id" db:"id"`
	Brand             *string    `json:"brand,omitempty" db:"brand"`
	Name              *string    `json:"name,omitempty" db:"name"`
	Category          *string    `json:"category,omitempty" db:"category"`
	ModelVariant      *string    `json:"model_variant,omitempty" db:"model_variant"`
	SellPackagingCost int        `json:"sell_packaging_cost" db:"sell_packaging_cost"`
	SellPostageCost   int        `json:"sell_postage_cost" db:"sell_postage_cost"`
	NewPrice          *int       `json:"new_price,omitempty" db:"new_price"`
	Enabled           *bool      `json:"enabled,omitempty" db:"enabled"`
	CreatedAt         *time.Time `json:"created_at,omitempty" db:"created_at"`
}

type TradedItem struct {
	ID                    int64      `json:"id" db:"id"`
	ProductID             *int64     `json:"product_id,omitempty" db:"product_id"`
	Storage               *int       `json:"storage" db:"storage"`
	ColorID               *int64     `json:"color_id,omitempty" db:"color_id"`
	BuyPrice              int        `json:"buy_price" db:"buy_price"`
	BuyShippingCost       int        `json:"buy_shipping_cost" db:"buy_shipping_cost"`
	BuyTransactionID      *int64     `json:"buy_transaction_id,omitempty" db:"buy_transaction_id"`
	BuyDate               *time.Time `json:"buy_date,omitempty" db:"buy_date"`
	SellPrice             *int       `json:"sell_price,omitempty" db:"sell_price"`
	SellPackagingCost     *int       `json:"sell_packaging_cost,omitempty" db:"sell_packaging_cost"`
	SellPostageCost       *int       `json:"sell_postage_cost,omitempty" db:"sell_postage_cost"`
	SellShippingCollected *int       `json:"sell_shipping_collected,omitempty" db:"sell_shipping_collected"`
	SellTransactionID     *int64     `json:"sell_transaction_id,omitempty" db:"sell_transaction_id"`
	SellDate              *time.Time `json:"sell_date,omitempty" db:"sell_date"`
	StatusID              int16      `json:"status_id" db:"status_id"`
	SourceLink            string     `json:"source_link" db:"source_link"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	ListingID             *int64     `json:"listing_id,omitempty" db:"listing_id"`
}

type TradeStatus struct {
	ID   int16  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type Listing struct {
	ID                  int64      `json:"id" db:"id"`
	ProductID           *int64     `json:"product_id,omitempty" db:"product_id"`
	Price               *int       `json:"price,omitempty" db:"price"`
	Valuation           int        `json:"valuation" db:"valuation"`
	Link                string     `json:"link" db:"link"`
	ConditionID         *int64     `json:"condition_id,omitempty" db:"condition_id"`
	ShippingCost        *int       `json:"shipping_cost,omitempty" db:"shipping_cost"`
	Title               string     `json:"title" db:"title"`
	Description         *string    `json:"description,omitempty" db:"description"`
	MarketplaceID       *int64     `json:"marketplace_id,omitempty" db:"marketplace_id"`
	Status              string     `json:"status" db:"status"`
	PublicationDate     *time.Time `json:"publication_date,omitempty" db:"publication_date"`
	SoldDate            *time.Time `json:"sold_date,omitempty" db:"sold_date"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	IsMyListing         bool       `json:"is_my_listing" db:"is_my_listing"`
	EligibleForShipping *bool      `json:"eligible_for_shipping,omitempty" db:"eligible_for_shipping"`
	SellerPaysShipping  *bool      `json:"seller_pays_shipping,omitempty" db:"seller_pays_shipping"`
	BuyNow              *bool      `json:"buy_now,omitempty" db:"buy_now"`
}

type Transaction struct {
	ID              int64     `json:"id" db:"id"`
	Date            time.Time `json:"date" db:"date"`
	Amount          int       `json:"amount" db:"amount"`
	TransactionType *int64    `json:"transaction_type,omitempty" db:"transaction_type"`
}

type TransactionType struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type Marketplace struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Link string `json:"link" db:"link"`
}

type Condition struct {
	ID    int64  `json:"id" db:"id"`
	Title string `json:"title" db:"title"`
}

type Color struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type ImageLink struct {
	ID        int64  `json:"id" db:"id"`
	URL       string `json:"url" db:"url"`
	ListingID int64  `json:"listing_id" db:"listing_id"`
}

type Economics struct {
	ID           int64 `json:"id" db:"id"`
	MinProfitSEK *int  `json:"min_profit_sek,omitempty" db:"min_profit_sek"`
	MinDiscount  *int  `json:"min_discount,omitempty" db:"min_discount"`
}

type TradedItemCandidate struct {
	Item          TradedItem
	EstimatedSell int
	ShippingCost  int
	TotalCost     int
	ShouldBuy     bool
}

type SearchTerm struct {
	ID            int64     `json:"id" db:"id"`
	Description   string    `json:"description" db:"description"`
	URL           string    `json:"url" db:"url"`
	MarketplaceID *int64    `json:"marketplace_id" db:"marketplace_id"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type CronJob struct {
	ID             int64     `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	CronExpression string    `json:"cron_expression" db:"cron_expression"`
	SearchTermIDs  []int64   `json:"search_term_ids" db:"search_term_ids"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type SearchCriteria struct {
	ID            int64     `json:"id" db:"id"`
	SearchTermID  int64     `json:"search_term_id" db:"search_term_id"`
	MarketplaceID *int64    `json:"marketplace_id,omitempty" db:"marketplace_id"`
	MaxPrice      *int      `json:"max_price,omitempty" db:"max_price"`
	MinCondition  *int64    `json:"min_condition,omitempty" db:"min_condition"`
	ExtraParams   string    `json:"extra_params" db:"extra_params"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type SearchTermWithCriteria struct {
	SearchTerm SearchTerm
	Criteria   []SearchCriteria
}

type ValuationType struct {
	ID      int16  `json:"id" db:"id"`
	Name    string `json:"name" db:"name"`
	Enabled bool   `json:"enabled" db:"enabled"`
}

type ProductValuationTypeConfig struct {
	ProductID       int64 `json:"product_id" db:"product_id"`
	ValuationTypeID int16 `json:"valuation_type_id" db:"valuation_type_id"`
	IsActive        bool  `json:"is_active" db:"is_active"`
}

type Valuation struct {
	ID              int64           `json:"id" db:"id"`
	ProductID       *int64          `json:"product_id,omitempty" db:"product_id"`
	ValuationTypeID *int16          `json:"valuation_type_id,omitempty" db:"valuation_type_id"`
	Valuation       int             `json:"valuation" db:"valuation"`
	Metadata        json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

type ValuationWithProduct struct {
	Valuation
	ProductName string `json:"product_name"`
}

type SearchHistory struct {
	ID              int64     `json:"id" db:"id"`
	SearchTermID    int64     `json:"search_term_id" db:"search_term_id"`
	SearchTermDesc  string    `json:"search_term_desc" db:"search_term_desc"`
	URL             string    `json:"url" db:"url"`
	ResultsFound    int       `json:"results_found" db:"results_found"`
	NewAdsFound     int       `json:"new_ads_found" db:"new_ads_found"`
	MarketplaceID   *int64    `json:"marketplace_id,omitempty" db:"marketplace_id"`
	MarketplaceName string    `json:"marketplace_name" db:"marketplace_name"`
	SearchedAt      time.Time `json:"searched_at" db:"searched_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type ScrapingRun struct {
	ID                 int64      `json:"id" db:"id"`
	StartedAt          time.Time  `json:"started_at" db:"started_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	Status             string     `json:"status" db:"status"`
	TotalAdsFound      int        `json:"total_ads_found" db:"total_ads_found"`
	TotalListingsSaved int        `json:"total_listings_saved" db:"total_listings_saved"`
	TotalGoodBuys      int        `json:"total_good_buys" db:"total_good_buys"`
	ErrorMessage       *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

type Conversation struct {
	ID            int64      `json:"id" db:"id"`
	ListingID     int64      `json:"listing_id" db:"listing_id"`
	MarketplaceID int64      `json:"marketplace_id" db:"marketplace_id"`
	Status        string     `json:"status" db:"status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type Message struct {
	ID             int64      `json:"id" db:"id"`
	ConversationID int64      `json:"conversation_id" db:"conversation_id"`
	Direction      string     `json:"direction" db:"direction"`
	Content        string     `json:"content" db:"content"`
	Status         string     `json:"status" db:"status"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty" db:"approved_at"`
	SentAt         *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type ConversationWithDetails struct {
	Conversation
	ListingTitle    string `json:"listing_title"`
	ListingPrice    *int   `json:"listing_price,omitempty"`
	MarketplaceName string `json:"marketplace_name"`
	PendingCount    int    `json:"pending_count"`
}
