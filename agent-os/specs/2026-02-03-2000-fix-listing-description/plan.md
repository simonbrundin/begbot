# Plan: Fix Listing Description Scraping Issue

## Problem
The current blocket scraping implementation saves malformed descriptions containing:
- Breadcrumb navigation text ("Du är härTorget/Elektronik och vitvaror...")
- UI elements and buttons ("Köp nuSkicka prisförslagTrygga affärer...")
- Page metadata ("Annonsens metadataDela-ikonDela annons...")
- Full HTML text content instead of actual ad description

## Root Cause
The `parseBlocketHTML` function extracts the `description` field from JSON-LD structured data, but this field contains the full rendered page text, not the actual ad description.

## Solution Options

### Option A: Use Blocket API Directly
- Use `https://blocket-api.se/v1/ad/recommerce?id={ad_id}` to get structured data
- Rate limited to 5 requests/second
- Returns clean, structured JSON with proper description field

### Option B: Fix HTML Parsing
- Extract ad description from specific HTML elements
- Requires understanding Blocket's HTML structure
- Less reliable as HTML can change

### Recommendation
**Use Blocket API directly** - cleaner, more reliable, properly structured data.

## Tasks
1. ~~Save spec documentation~~ ✅ COMPLETED
2. ~~Analyze Blocket API response structure~~ ✅ COMPLETED
3. ~~Create API client function for Blocket API~~ ✅ COMPLETED
4. ~~Update `fetchBlocketAdsFromURL` to use API~~ ✅ COMPLETED
5. ~~Handle API rate limiting (5 req/s)~~ ✅ COMPLETED
6. ~~Test with sample listings~~ ✅ COMPLETED

---

## Spec Status: ✅ DONE

**Completed:** 2026-02-03

---

## Summary

Fixed the listing description scraping issue by implementing proper Blocket API integration:

1. **Added rate limiting** (`waitForRateLimit` method) to respect the 5 req/s limit
2. **Fixed ad ID extraction** to support both `/item/` and `/annons/` URL patterns
3. **Added unit tests** for rate limiting and ad ID extraction
4. **Verified** API integration works correctly with real Blocket API calls

## Files Modified

- `internal/services/marketplace.go` - Added rate limiting and fixed regex
- `internal/services/marketplace_test.go` - New test file

## Verification

- All unit tests pass
- Rate limiting verified to enforce 5 requests/second
- API integration tested with real Blocket API
