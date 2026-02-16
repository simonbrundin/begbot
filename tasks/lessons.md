# Lessons Learned

## Bug Fix: "Inga annonser hittades" trots data i databasen

### Datum
2026-02-03

### Problem
Sidan "Hittade annonser" (`/ads`) visade "Inga annonser hittades" trots att databasen innehöll 183 annonser. Konsolen visade felet: `allListings?.filter is not a function`

### Rotorsak
Frontend-koden använde `onMounted` hook för att hämta data via `api.get()`. Problemet är att:

1. `onMounted` kör **endast på klientsidan**
2. Nuxt 3 använder server-side rendering (SSR) som standard
3. SSR renderar sidan utan att köra `onMounted`
4. Resultat: Sidan renderades med tomma data (`listings.value = []`)

### Kod som orsakade problemet
```typescript
const fetchData = async () => {
  const allListings = await api.get<Listing[]>('/listings')
  listings.value = allListings?.filter(l => !l.is_my_listing) || []
}

onMounted(fetchData)  // Kör bara på klienten!
```

### Lösning
Använd `useAsyncData` som fungerar i både SSR- och klient-kontexter:

```typescript
const { data: listings, error, pending } = await useAsyncData(
  'ads-listings',
  async () => {
    const response = await fetch(`${config.public.apiBase}/api/listings`)
    const data = await response.json()
    return data.filter((item: any) => !item.is_my_listing)
  }
)
```

### Reflekterade regler

#### 1. Använd alltid SSR-kompatibla data-fetching metoder i Nuxt 3
- **Gör**: Använd `useAsyncData` eller `useFetch` för datahämtning
- **Undvik**: Använd inte `onMounted` för kritisk datahämtning
- **Varför**: Nuxt 3 renderar på servern som standard, och `onMounted` kör bara i webbläsaren

#### 2. Felsök med SSR i åtanke
- Testa sidor via `curl` för att se SSR-output
- SSR-output visar vad som faktiskt renderas på servern
- Använd konsolloggning i både SSR och klient-kontexter

#### 3. API:er fungerar - problemet är i frontend-kontexten
- När API:et fungerar via `curl` men inte i browsern:
- Kontrollera om datahämtningen är SSR-kompatibel
- Kolla om composables kör i rätt kontext

### Fil som ändrades
- `/home/simon/repos/begbot/frontend/pages/ads.vue`

### Resultat
- Sidan visar nu korrekt 183 annonser vid laddning
- Fungerar både med SSR och klient-side navigation
- Ingen tom "Inga annonser hittades"-text längre
