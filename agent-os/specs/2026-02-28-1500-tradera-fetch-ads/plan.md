# Plan: Hämta annonser från Tradera + email-policy

1) Save spec documentation (this file + shape/standards/references/visuals)

2) Problemställning
- Vi vill säkerställa att Tradera-annonser hämtas korrekt både via SOAP/API och via HTML-scraping.
- Viktigt: systemet får aldrig skicka trading-rekommendationer för en Tradera-annons enbart baserat på utgångspris (auktion). Rekommendationer/e‑post får skickas endast när annonsen har ett uttryckligt "Köp nu" (buy-now) pris.

3) Bakgrund / nuläge
- Majoriteten är redan implementerat: se `internal/services/marketplace.go` (fetchTraderaAds, ParseTraderaDoc) och `internal/marketplaces/tradera.go` (SOAP client).
- E‑post-logiken är redan centraliserad i `internal/services/bot.go:shouldSendTradingRuleEmail` där det finns en Tradera-specifik check som skippar mail om `ad.HasBuyNow == false`.
- Risk: när Tradera hämtas via SOAP/API så saknas fält som `HasBuyNow`/`BuyNowPrice` i nuvarande `marketplaces.RawAd` vilket gör att e‑post kan felaktigt undertryckas eller skickas.

4) Hög-nivå lösning
- Se till att varje väg som returnerar Tradera-annonser (API och scraping) sätter dessa fält konsekvent: `HasBuyNow`, `BuyNowPrice`, `CurrentPrice`.
- If API lacks buy-now info, enrichera API-resultatet genom att anropa GetItem (ad-detaljer) för de annonser där det behövs — respektera Tradera API rate-limits.
- Anpassa konverteringskod så att interna `RawAd` som konsumeras av bot/logik alltid har korrekt `HasBuyNow` värde.

5) Tasks (implementation, i ordning)
- Task 1 (mandatory): Save spec documentation to this folder. (done)
- Task 2: Add buy-now fields to marketplace-level ad structs so both packages share the contract
  - Files: `internal/marketplaces/marketplace.go`, `internal/services/marketplace.go`
  - Fields: `HasBuyNow bool`, `BuyNowPrice float64`, `CurrentPrice float64`
- Task 3: Populate buy-now details for Tradera SOAP client
  - Inspect SearchResponse/SearchItem for buy-now fields, populate CurrentPrice/BuyNowPrice when possible
  - If Search doesn't include buy-now, call GetItem for each item (or for subset) to get definitive data
  - Add rate-limit handling and batching to avoid exceeding Tradera API limits
- Task 4: Ensure convert/adapter preserves buy-now fields
  - Update `convertMarketplacesRawAds` to copy new fields
  - If enrichment is async/extra-call, ensure MarketplaceService uses enriched ads for downstream processing
- Task 5: Tests
  - Unit tests for ParseTraderaDoc: ensure it detects buy-now vs auction and sets HasBuyNow
  - Unit tests for Tradera SOAP client: mock Search + GetItem to assert enrichment logic and rate-limit behaviour
  - Integration/e2e test: scenario where API path is used and an ad with buy-now results in email being sent; scenario where auction-only results in no email
- Task 6: Update bot/email logic (small review)
  - Verify `shouldSendTradingRuleEmail` uses `ad.HasBuyNow` (already implemented), and ensure no other code path overrides it.
- Task 7: Acceptance + rollout
  - Add monitoring/logging when emails are skipped for Tradera (reason: auction-only)
  - Deploy behind feature-flag if desired; otherwise run canary tests and monitor logs

6) Acceptance criteria
- Tradera ads fetched via API and scraping have `HasBuyNow` correctly set.
- Bot only sends trading rule emails for Tradera when `HasBuyNow == true`.
- Tests cover positive and negative cases (buy-now present / absent).

7) Verification steps for reviewer
 - Run unit tests: `go test ./...` and confirm new tests pass
 - Run `./bin/begbot` (or suitable dev run) against sample Tradera HTML fixtures and mock API responses
 - Check logs for lines: "Skipping email for auction-only Tradera listing" and "Sent trading rule email for listing"

8) Risks
 - Extra API calls to GetItem may hit Tradera rate limits -> implement caching and backoff
 - Parsing HTML may break if Tradera markup changes; keep tests with representative fixtures
