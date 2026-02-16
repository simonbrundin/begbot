# Shaping Notes: Listing Description Fix

## Context
The blocket scraper is saving garbled text instead of clean ad descriptions. Database query shows descriptions contain full page HTML text.

## Current Implementation
- **File**: `internal/services/marketplace.go`
- **Function**: `parseBlocketHTML` (lines 299-410)
- **Issue**: Extracts `item.Item.Description` from JSON-LD which contains full page text

## Blocket API Reference
Base URL: `https://blocket-api.se`

Endpoints:
- Search: `GET /v1/search?query={query}`
- Get Ad: `GET /v1/ad/recommerce?id={ad_id}`
- Rate limit: 5 requests/second

## Decision
Use Blocket API directly for fetching ad details. The API provides:
- Clean description field
- Proper price structure
- Condition information
- Shipping details

## Changes Required
1. Add `fetchBlocketAdFromAPI` function
2. Update search to collect ad IDs
3. Fetch full details via API
4. Replace/augment HTML parsing with API data

## Scope
Minimal - only fix description extraction. Keep existing shipping cost parsing from HTML as fallback.
