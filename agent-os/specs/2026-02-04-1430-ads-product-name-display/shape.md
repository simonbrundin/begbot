# Shaping: Visa produktnamn på annonser

## Beskrivning
På sidan `/ads` ska varje annons visa vilket produkt den är kopplad till (från `products` tabellen). Detta gör det lättare att se vilken produkt varje annons representerar.

## Scope

### Ingår
- Uppdatera frontend-komponenten för att alltid visa produktnamn
- Säkerställa att produktdata visas tydligt i annonskorten
- Behålla befintlig layout och styling

### Ingår inte
- Ändringar i backend API (data returneras redan)
- Ändringar i databasstruktur
- Nya funktioner eller filter

## Beslut

1. **Visning:** Visa produktnamn tydligt på varje annonskort
2. **Placering:** Under eller bredvid titeln för god läsbarhet
3. **Format:** "Produkt: {brand} {name}" för tydlighet
4. **Fallback:** Om ingen produkt är kopplad, visa "Okänd produkt"

## Kontext
- API:et returnerar redan `Product` objektet via `GetListingsWithProfit`
- Frontend-typen `ListingWithDetails` innehåller redan `Product`
- Nuvarande kod visar produktnamn endast som fallback när titel saknas

## Förväntat resultat
Varje annonskort på `/ads` visar tydligt vilken produkt annonsen är kopplad till.
