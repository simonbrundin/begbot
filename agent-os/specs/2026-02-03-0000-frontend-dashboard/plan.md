# Plan: Frontend Dashboard

## Overview

Build a Nuxt 3 + Vue 3 dashboard for managing all inventory, listings, products, and analytics. Full CRUD operations with Supabase Auth authentication.

## Pages

1. **Inventory** - Traded items (buy/sell tracking, profit calculation)
2. **My Listings** - Listings with `is_my_listing = true`
3. **Product Catalog** - Products with enable/disable toggle
4. **Analytics** - Profit reports, inventory value, sales stats
5. **Scraping Settings** - Search terms management
6. **Transactions** - Financial transactions

## Tasks

### Task 1: Save spec documentation
- [x] Save this plan
- [x] Save shape.md
- [x] Save standards.md
- [x] Save references.md

### Task 2: Initialize Nuxt 3 project ✅ COMPLETED
- Create `frontend/` directory with Nuxt 3
- Install dependencies (Tailwind CSS, Supabase client, etc.)
- Configure nuxt.config.ts

### Task 3: Set up Supabase Auth ✅ COMPLETED
- Install `@nuxtjs/supabase`
- Configure Supabase URL and anon key from env
- Create auth pages (login, callback)
- Add auth middleware

### Task 4: Create database types ✅ COMPLETED
- Generate types from existing Go models
- Add to `types/database.ts`

### Task 5: Build API routes in Go backend ✅ COMPLETED
- Add REST endpoints for all CRUD operations
- Endpoints: `/api/inventory`, `/api/listings`, `/api/products`, `/api/transactions`, `/api/search-terms`

### Task 6: Create layouts and navigation ✅ COMPLETED
- Default layout with sidebar navigation
- Responsive design with Tailwind

### Task 7: Build Inventory page ✅ COMPLETED
- Table view of all traded items
- Status badges (potential, purchased, in_stock, listed, sold)
- Profit calculation display
- Add/Edit/Delete items
- Filter by status

### Task 8: Build My Listings page ✅ COMPLETED
- Grid/card view of own listings
- Edit price, status, description
- Link to marketplace

### Task 9: Build Product Catalog page ✅ COMPLETED
- List all products
- Enable/disable products
- Add new products

### Task 10: Build Analytics page ✅ COMPLETED
- Total inventory value
- Total profit earned
- Items by status chart
- Recent sales

### Task 11: Build Scraping Settings page ✅ COMPLETED
- List search terms
- Add/Edit/Delete search terms
- Toggle active status

### Task 12: Build Transactions page ✅ COMPLETED
- List all transactions
- Add new transactions
- Transaction type filter

### Task 13: Add global styles ✅ COMPLETED
- Configure Tailwind CSS
- Create app.css for custom styles

### Task 14: Write tests ✅ COMPLETED
- Created test suite in `tests/utils.test.ts`
- 5 tests for formatCurrency, formatDate, calculateProfit
- All tests passing with Vitest

### Task 15: Verify and lint ✅ COMPLETED
- `npm run build` succeeds (production build works)
- Added `npm run test` script
- Vitest configured with `@nuxt/test-utils`
- 5 unit tests passing

---

## Spec Status: ✅ DONE

Completed: 2026-02-03

All tasks complete. Frontend dashboard is fully functional with:
- 7 pages (Inventory, Listings, Products, Transactions, Analytics, Scraping, Ads)
- Full CRUD operations via REST API
- Supabase Auth integration
- Tailwind CSS styling
- Vitest unit tests
