# Shape: Search Terms Table

## Scope

Create a system for storing and managing search terms that the bot uses to discover new products on marketplaces. The system must support:
- Storing complete URLs with pre-configured filters
- Active/inactive toggle
- Integration with existing scraping infrastructure
- Simple marketplace_id foreign key

## Context

This feature extends the existing marketplace scraping system. Currently, `phones.yaml` contains hardcoded URLs for iPhone searches. The goal is to make this configurable via the database.

The existing architecture has:
- `marketplaces` table with marketplace definitions
- `listings` table for storing discovered ads
- `MarketplaceService` with working fetch methods
- `RawAd` struct for ad data

## Decisions

### Simple URL Storage

Store the complete URL directly instead of building it from criteria. The user copies the URL from their web browser where they have already configured filters. This is simpler because:
- No need to build URLs programmatically
- No complex criteria tables
- User has full control over filters via web UI
- Easier to understand and maintain

### Single Table Schema

Use one simple table with:
- `name` - human-readable label
- `url` - complete URL with filters (copy from browser)
- `marketplace_id` - foreign key to marketplaces table
- `is_active` - toggle on/off

### Duplicate Handling

Check for existing listings by `link` before inserting.

## Constraints

- Must use existing `marketplaces` table foreign key
- Must follow currency storage standard (integers in SEK öre)
- Must integrate with existing `MarketplaceService`
- Must not break existing functionality
