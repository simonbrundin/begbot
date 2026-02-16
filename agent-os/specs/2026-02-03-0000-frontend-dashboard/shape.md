# Shaping Decisions

## Scope Decision: Full CRUD Dashboard

User wants complete control over all database data. Not limited to schema structure - free to present data in intuitive ways.

## Tech Stack

- **Framework**: Nuxt 3 (Vue 3)
- **Styling**: Tailwind CSS
- **Backend**: Go (existing) + REST API
- **Auth**: Supabase Auth
- **Database**: PostgreSQL via Supabase

## Key Design Decisions

### Page Structure
- Single-page application with sidebar navigation
- Modal dialogs for editing data
- Toast notifications for actions

### Data Presentation
- Tables for list views (inventory, transactions)
- Cards for listings and products
- Charts for analytics

### API Design
- REST endpoints in Go backend
- Reuse existing database functions in postgres.go
- JSON response format matching Vue types

## Database Schema for Reference

```
products, traded_items, listings, transactions, search_terms,
colors, conditions, trade_statuses, marketplaces, image_links
```

## User Requirements

1. Edit all data, not just view
2. Supabase Auth for security
3. All pages: Inventory, My Listings, Product Catalog, Analytics, Scraping Settings
4. Free to reorganize data presentation
