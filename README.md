# Begbot

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
         1. Värdering från egen databas
            1. Sortera sålda varor på pris och gör en graf med x-axel pris och
               y-axel dagar den låg ute till försäljning.
            2. Hitta grafen k-värde och välj priset för vår valda
               försäljningstid (tex 14 dagar)
            3. Är priset dyrare avbryt inköpsprocess
         2. Traderas värderingsverktyg
            - https://www.tradera.com/valuation
         3. Sålda annonser
            1. Marketplace
            2. Ebay
         4. Nypris
            1. LLM tar fram nypris
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
