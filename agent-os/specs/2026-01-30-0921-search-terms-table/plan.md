# Plan: Search Terms Table for Marketplace Scraping

## Overview

Create a database table structure for storing search terms that can be used to search across multiple marketplaces (Blocket, Tradera). The search terms will have filtering criteria, active/inactive status, and will integrate with the existing marketplace scraping service to discover new listings.

## Tasks

1. ~~**Save spec documentation**~~ ✅ COMPLETED
   - Document scope, standards, and references

2. ~~**Design and create database schema**~~ ✅ COMPLETED
   - Create `search_terms` table with name, url, marketplace_id, is_active, timestamps
   - Create indexes for efficient querying

3. ~~**Add Go models**~~ ✅ COMPLETED
   - Add `SearchTerm` struct in internal/models/models.go
   - Add `SearchCriteria` struct in internal/models/models.go
   - Add marketplace relationship fields

4. ~~**Add database methods**~~ ✅ COMPLETED
   - Add `SaveSearchTerm` method in internal/db/postgres.go
   - Add `GetActiveSearchTerms` method
   - Add `GetSearchCriteriaByTermID` method
   - Add `UpdateSearchTermStatus` method

5. ~~**Create search term management service**~~ ✅ COMPLETED
   - Create internal/services/search_terms.go
   - Add `SearchTermService` struct
   - Add CRUD operations for search terms
   - Add method to build marketplace URLs from criteria

6. ~~**Integrate with marketplace service**~~ ✅ COMPLETED
   - Extend MarketplaceService to accept search term criteria
   - Add method to fetch ads with custom filters
   - Handle duplicate detection before saving listings

7. ~~**Create CLI commands for search term management**~~ ✅ COMPLETED
   - Add cmd/searchterms/main.go for managing search terms
   - Add command to list active search terms
   - Add command to add new search term
   - Add command to run scraping for all active terms

8. ~~**Write unit tests**~~ ✅ COMPLETED
   - Test database CRUD operations
   - Test URL building logic
   - Test duplicate detection

## Technical Design

### Database Schema

```sql
CREATE TABLE search_terms (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    query TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE search_criteria (
    id SERIAL PRIMARY KEY,
    search_term_id INTEGER REFERENCES search_terms(id) ON DELETE CASCADE,
    marketplace_id SMALLINT REFERENCES marketplaces(id),
    max_price INTEGER,
    min_condition SMALLINT,
    shipping_types TEXT,
    sort_order TEXT DEFAULT 'PUBLISHED_DESC',
    extra_params TEXT  -- JSON for additional filters
);
```

### URL Building Strategy

Build marketplace URLs dynamically based on criteria:
- Blocket: Parse existing URL pattern, replace/add query params
- Tradera: Similar pattern matching

### Duplicate Detection

Before saving new listings, check if link already exists in listings table using source_link unique constraint or index.

## Definition of Done

- [x] Database schema created with proper constraints
- [x] Go models added and compilable
- [x] Database CRUD methods implemented
- [x] Search term service with URL building
- [x] Integration with existing marketplace scraping
- [x] CLI commands for management
- [x] Unit tests passing
- [x] Documentation updated

---

## Spec Status: ✅ DONE

Completed: 2026-02-01

