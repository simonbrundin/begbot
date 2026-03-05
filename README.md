# Begbot

## Tradera Annonshantering

### Hämtningsflöde (Annonslistor)

När begbot hämtar annonser från Tradera används följande fallback-kedja:

```
1. Tradera SOAP API (prioriterad)
   ↓ (om API misslyckas eller saknas)
2. Direkt scraping (utan proxy)
   ↓ (om blockerad eller misstänkt)
3. Evomi Scraper (partner)
   ↓ (om Evomi misslyckas)
4. Proxy-baserad fetch (sista utväg)
```

Se `internal/services/marketplace.go:203` för implementationsdetaljer.

### Enrichment-flöde (Detaljer per annons)

För varjeTradera-annons som ska sparas finns följande flöde:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        TRADERA ENRICHMENT FLOW                              │
├─────────────────────────────────────────────────────────────────────────────┤
│  1. [START] Processing ad: https://www.tradera.com/item/xxx                │
│      → Initial price: XXX SEK, Marketplace: tradera                         │
│                                                                             │
│  2. [LLM] Extract product info                                             │
│      → Manufacturer, Model, Category, Storage                              │
│                                                                             │
│  3. [SHIPPING] Determine shipping costs                                     │
│      → Use Blocket API if available, else LLM estimate                    │
│                                                                             │
│  4. [VALIDATION] Match against existing products                           │
│      → Create new product if not found                                     │
│                                                                             │
│  5. [INTACT] Evaluate product condition                                    │
│      → Skip if damaged or has issues                                       │
│                                                                             │
│  6. [TRADERA] Enrichment (if SaveOnlyBuyNow is enabled)                  │
│      ┌────────────────────────────────────────────────────────────────┐    │
│      │  a) API Enrichment (GetItem)                                  │    │
│      │     → Attempt 1/3: Try Tradera API                           │    │
│      │        ├─ SUCCESS: Found buy-now → Continue                  │    │
│      │        └─ FAIL (429): Rate limited → Backoff & Retry         │    │
│      │     → Attempt 2/3: Retry after backoff                       │    │
│      │     → Attempt 3/3: Retry after backoff                      │    │
│      │                                                              │    │
│      │  b) Fallback: Scrape ad page directly                        │    │
│      │     (if all API attempts failed)                             │    │
│      │     → Parse HTML for buy-now price                           │    │
│      │        ├─ SUCCESS: Found buy-now → Continue                 │    │
│      │        └─ FAIL: No buy-now → Skip ad (auction only)         │    │
│      │                                                              │    │
│      │  c) Skip ad if:                                              │    │
│      │     - No buy-now price found                                 │    │
│      │     - SaveOnlyBuyNow disabled                                │    │
│      │     - No tradera client available                            │    │
│      └────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  7. [PRICE] Final price validation                                         │
│      → Skip if price <= 0                                                  │
│                                                                             │
│  8. [DB] Save listing to database                                          │
│      → Save images, valuations                                             │
│                                                                             │
│  9. [SUCCESS] Listing saved                                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Loggningsnivåer

Alla steg loggas med tydliga prefix:

- `[START]` - Början av bearbetning
- `[LLM]` - LLM-relaterade anrop
- `[SHIPPING]` - Fraktberäkning
- `[VALIDATION]` - Produktvalidering
- `[INTACT]` - Skick-utvärdering
- `[TRADERA]` - Tradera-enrichment
- `[PRICE]` - Prisvalidering
- `[DB]` - Databasoperationer
- `[SUCCESS]` - Framgångsrik sparning

### Konfigurationsalternativ

I `config.yaml` under `tradera`:

```yaml
tradera:
  save_only_buy_now: true # Spara endast annonser med köp nu-pris
  enrich_timeout_seconds: 5 # Timeout för enrichment (sekunder)
  enrich_max_retries: 3 # Antal försök vid API-anrop
  enrich_backoff_seconds: 2 # Bas-backofftid vid rate limiting
```

### Exempel på loggutskrift

```
[INFO] ========================================
[INFO] [START] Processing ad: https://www.tradera.com/item/123/456/test
[INFO] [INFO] Marketplace: tradera, Initial price: 0.00 SEK
[INFO] [INFO] Title: Apple Mac Mini M4 16GB
[INFO] [LLM] Extracted product info: Manufacturer="Apple", Model="Mac Mini M4"
[INFO] [SHIPPING] Final: cost=59 SEK, insurance=0 SEK
[INFO] [VALIDATION] Validating listing against existing products...
[INFO] [VALIDATION] Matched to product ID=42: Apple Mac Mini M4
[INFO] [INTACT] Evaluating product condition...
[INFO] [INTACT] Result: isIntact=true, issues=[]
[INFO] [INTACT] Product is intact - continuing
[INFO] [TRADERA] Starting enrichment flow for: https://www.tradera.com/item/123/456/test
[INFO] [TRADERA] Ad has no buy-now price, checking SaveOnlyBuyNow policy
[INFO] [TRADERA] SaveOnlyBuyNow is enabled, attempting enrichment
[INFO] [TRADERA] Attempting API enrichment (GetItem)
[INFO] [TRADERA] API enrichment attempt 1/3
[WARNING] [TRADERA] Rate limited (429) on attempt 1/3, backing off 2s
[INFO] [TRADERA] API enrichment attempt 2/3
[INFO] [TRADERA] API enrichment succeeded on attempt 2
[INFO] [TRADERA] API enrichment SUCCESS - buy-now: 5990.00 SEK, current: 4990.00 SEK
[INFO] [PRICE] Final price for listing: 5990 SEK
[INFO] [DB] Saving listing to database...
[INFO] [DB] Listing saved successfully (ID: 1234)
[INFO] [DB] Saving 4 image links...
[INFO] [DB] Image links saved successfully
[INFO] [SUCCESS] Listing saved: Apple Mac Mini M4 at 5990 SEK (valuation: 6500 SEK)
```

## Secrets Management

### Configuration

All sensitive configuration is stored in environment variables. The application
reads from:

1. **`.env`** file (local development, gitignored)
2. **Environment variables** (Kubernetes production)

### Safe to Commit

- `config.yaml` - Contains no secrets, only placeholders
- `.env.example` - Contains empty placeholders with TODO comments

### Not Committed

- `.env` - Contains actual secrets (in .gitignore)
- `.env.local` - Local overrides (in .gitignore)

### Kubernetes Production

Secrets are managed via **Vault** and **External Secrets Operator**:

1. Secrets stored in Vault at path `prod/begbot`
2. ExternalSecret resource (`base/externalsecret.yaml`) syncs to Kubernetes
   Secret
3. Deployment injects as environment variables:

```yaml
# base/deployment.yaml
envFrom:
  - configMapRef:
      name: begbot-config
  - secretRef:
      name: begbot-secrets
```

### Setup for Local Development

1. Copy `.env.example` to `.env`
2. Fill in the actual secrets
3. Run `go run ./cmd/main.go`

**Note:** For production, add secrets to Vault:

```bash
vault kv put prod/begbot database_password=xxx llm_api_key=xxx smtp_username=xxx smtp_password=xxx smtp_from=xxx
```

## Specs att köra

- [x] Säker lagring - Lagra känslig information på ett säkert sätt.
- [ ]
- [x] Frontend - skapa en frontend för min applikation där jag kan se vad jag
      har till salu. alla annonser som sparats mm. egentligen allt i min databas
      men i ett gui. jag vill att det ska vara byggt med nuxt och vue för det
      har jag använt många gånger.
- [x] Refactor för att förenkla
- [x] Opencode har inte rättigheter att göra förändringar i databasen utan att
      be om lov
- [x] Lägg upp deployment på samma sätt som i plan

## Inköpsprocess

1. Hämta alla nya annonser på en specifik vara
   1. Spara ner dom till en array
   2. Jämför med tidigare cache
   3. Släng länkar som redan finns i cache
2. Loopa igenom varje vara i cache
   1. Spara ner produkten till egen databas
      - Information i databasen
        - Tillverkare
        - Modell
        - Lagring
        - Skick
        - Fraktkostnad
        - Annonstext
        - Länk
        - Pris
        - Marknadsplats
        - Annonsdatum
        - Bildlänkar
      1. Identifiera produktinformation
         1. Skicka länk till llm och definera information vi vill ha i databasen
      2. Säkerställ att det är rätt produkt (tex ingen mobilskal)
      3. Säkerställ att produkten är hel
         1. Mindre repor okej i övrigt ska de vara utan fel eller skador
         - (OM hel)
           1. Skapa SQL-query
           2. Kör query
         - (OM trasig) > return
   2. Värdera
      1. Samla in värderingar
         1. Egen värdering
         1. Värdering från egen databas
            1. Sortera sålda varor på pris och gör en graf med x-axel pris och
               y-axel dagar den låg ute till försäljning.
            2. Hitta grafen k-värde och välj priset för vår valda
               försäljningstid (tex 14 dagar)
            3. Är priset dyrare avbryt inköpsprocess
         1. Traderas värderingsverktyg
            - https://www.tradera.com/valuation
         1. Sålda annonser
            1. Marketplace
            2. Ebay
      2. Sammanställ värdering
         - LLM sammanställer till ett pris och en säkerhetsprocent
   3. Frakt
      1. Säkerställ fraktmöjlighet a. Läs annonstext om det står något om frakt
      2. Räkna ut fraktkostnad
   4. Räkna ut inköpspris
   5. Räkna ut vinst
   6. Bedöm säkerhet för hur snabbt den säljs
   7. Skicka erbjudande för granskning ifall vinsten är bra nog
   8. Köp
      - Budsida
        - Lägg maxbud när det är 30 sek kvar
      - Köpsida
