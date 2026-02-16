# Plan: Visa produktnamn på annonser

## Översikt
Lägg till visning av produktnamn (från `products` tabellen) på varje annons i `/ads` sidan.

## Struktur
```
agent-os/specs/2026-02-04-1430-ads-product-name-display/
├── plan.md           # Denna fil
├── shape.md          # Scope och beslut
├── standards.md      # Tillämpliga standarder
├── references.md     # Referenser till kod
└── visuals/          # (tom - inga mockups behövs)
```

## Tasks

### Task 1: Spara spec-dokumentation
**Status:** ✅ Klar
**Fil:** Skapa mappstruktur och alla spec-filer

### Task 2: Uppdatera ads.vue för att visa produktnamn
**Status:** ✅ Klar
**Fil:** `frontend/pages/ads.vue`
**Ändringar:**
- Lägg till visning av produktnamn på varje annonskort
- Visa "Produkt: {brand} {name}" under befintlig titel
- Hantera fall där produkt saknas (visa "Okänd produkt")
- Behåll nuvarande styling och layout
- **Uppdatering:** La till bindestreck mellan brand och name

**Specifika ändringar:**
- Efter rad 100 (där titel visas): Lägg till ny `<p>` tagg för produktnamn
- Använd klass: `text-sm text-slate-400`
- Format: "Produkt: {{ item.Product.brand }} - {{ item.Product.name }}"
- Fallback: Om `item.Product` är null, visa "Okänd produkt"

### Task 3: Verifiera ändringarna
**Status:** ✅ Klar
**Steg:**
1. ✅ Kontrollera att sidan `/ads` laddas utan fel
2. ✅ Verifiera att varje annons visar produktnamn
3. ✅ Kontrollera att fallback fungerar för annonser utan produkt
4. ✅ Säkerställa att styling är konsekvent med resten av UI

## Verifieringskriterier
- [x] Alla annonser visar produktnamn (eller "Okänd produkt")
- [x] Produkten visas med format: "Produkt: {brand} - {name}"
- [x] Styling följer befintligt tema (text-slate-400, text-sm)
- [x] Ingen påverkan på befintlig funktionalitet
- [x] Svensk text används ("Produkt:" inte "Product:")

## Implementation Notes
- Ingen ändring behövs i backend - data finns redan
- Använd befintlig `item.Product` från `ListingWithDetails` typen
- Följ befintlig kodstil i `ads.vue`

## Sammanfattning
✅ **Alla tasks klara**

Implementerat i `frontend/pages/ads.vue`:
- Produktnamn visas nu på varje annonskort
- Format: "Produkt: {brand} - {name}"
- Fallback: "Okänd produkt" om ingen produkt är kopplad
- Bindestreck mellan brand och name för tydligare separation
- Styling: text-sm text-slate-400 för konsekvent utseende
