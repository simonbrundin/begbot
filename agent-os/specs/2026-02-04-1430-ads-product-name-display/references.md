# Referenser

## Frontend

### Nuvarande implementation
- **Fil:** `frontend/pages/ads.vue`
- **Rad 95-101:** Nuvarande logik för att visa titel/produkt
  ```vue
  <p v-if="item.Listing.title" class="font-medium text-slate-100">
    {{ item.Listing.title }}
  </p>
  <p v-else-if="item.Product" class="font-medium text-slate-100">
    {{ item.Product.brand }} {{ item.Product.name }}
  </p>
  <p v-else class="text-slate-500">Okänd produkt</p>
  ```

### Typer
- **Fil:** `frontend/types/database.ts`
- **Interface:** `ListingWithDetails` (rad 124-130)
  - Innehåller `Product: Product | null`
  - Product har `brand: string | null` och `name: string | null`

## Backend

### API Endpoint
- **Fil:** `cmd/api/main.go`
- **Funktion:** `getListings` (rad 176-196)
- **Databas-funktion:** `GetListingsWithProfit` (rad 178)

### Databas
- **Fil:** `internal/db/postgres.go`
- **Funktion:** `GetListingsWithProfit` (rad 903-943)
  - Hämtar redan produktdata (rad 934-938)
  - Returnerar `ListingWithProfit` som innehåller `Product *models.Product`
