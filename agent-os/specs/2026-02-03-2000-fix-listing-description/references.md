# References: Listing Description Fix

## Blocket API
- API Documentation: https://blocket-api.se/api-reference/
- Swagger UI: https://blocket-api.se/swagger
- Repository: https://github.com/dunderrrrrr/blocket_api

## Current Code
- `internal/services/marketplace.go` - Existing marketplace scraping

## Database Schema
- `listings` table - stores `title`, `description`, `price`, etc.

## Related Files
- `internal/models/models.go` - `Listing` struct with `Description` field
- `internal/config/config.go` - `BlocketConfig` for API settings
