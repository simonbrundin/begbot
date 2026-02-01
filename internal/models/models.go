package models

import (
	"time"
)

type Product struct {
	ID                int64     `json:"id" db:"id"`
	Brand             string    `json:"brand" db:"brand"`
	Name              string    `json:"name" db:"name"`
	Category          string    `json:"category" db:"category"`
	ModelVariant      string    `json:"model_variant" db:"model_variant"`
	SellPackagingCost int       `json:"sell_packaging_cost" db:"sell_packaging_cost"`
	SellPostageCost   int       `json:"sell_postage_cost" db:"sell_postage_cost"`
	Enabled           bool      `json:"enabled" db:"enabled"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
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
	ID              int64      `json:"id" db:"id"`
	ProductID       *int64     `json:"product_id,omitempty" db:"product_id"`
	Price           *int       `json:"price,omitempty" db:"price"`
	Link            string     `json:"link" db:"link"`
	ConditionID     *int64     `json:"condition_id,omitempty" db:"condition_id"`
	ShippingCost    *int       `json:"shipping_cost,omitempty" db:"shipping_cost"`
	Description     string     `json:"description" db:"description"`
	MarketplaceID   *int64     `json:"marketplace_id,omitempty" db:"marketplace_id"`
	Status          string     `json:"status" db:"status"`
	PublicationDate *time.Time `json:"publication_date,omitempty" db:"publication_date"`
	SoldDate        *time.Time `json:"sold_date,omitempty" db:"sold_date"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	IsMyListing     bool       `json:"is_my_listing" db:"is_my_listing"`
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
