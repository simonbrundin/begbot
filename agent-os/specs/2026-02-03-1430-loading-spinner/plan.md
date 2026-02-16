# Loading Spinner - Implementation Plan

## Feature Overview
Skapa en global loading spinner som visas på varje sida medan applikationen väntar på svar från API:et. Spinnern ska vara synlig och tydlig utan att blockera hela UI:t.

## Scope
- **Ny funktion**: Vi har ingen loader/spinner idag
- **Tech stack**: Nuxt 3, Vue 3, Tailwind CSS v6
- **API-klient**: Använder `$fetch` för API-anrop

## Architecture Decisions

### 1. Global Loading State
Använd Pinia store för global loading state istället för lokala `loading` refs på varje sida. Detta ger:
- Konsistent beteende över hela appen
- Möjlighet att visa spinner även för inline/API-anrop
- Enklare att hantera multiple samtidiga requests

### 2. Spinner Placement
Spinner ska visas:
- Globalt i layout (overlay över hela sidan)
- Med semi-transparent bakgrund
- Centrerad med tydlig animation
- Ej blockera hela UI:t (användaren ska fortfarande kunna interagera med redan laddad data)

### 3. Visual Design
- Tailwind-baserad animation
- Matchar dark theme (slate-900 bakgrund)
- Använder emerald-400 som accentfärg (samma som Begbot-loggan)
- Minimalistisk, inte för distraherande

## Implementation Tasks

### Task 1: Save spec documentation
Skapa all dokumentation i `agent-os/specs/YYYY-MM-DD-HHMM-loading-spinner/`

### Task 2: Create Pinia store for loading state
- Skapa `stores/loading.ts`
- Hantera counter för aktiva requests
- Exponera `isLoading` computed property

### Task 3: Create LoadingSpinner component
- Skapa `components/LoadingSpinner.vue`
- Tailwind-baserad CSS-animation
- Centrerad overlay med `fixed` position
- Semi-transparent bakgrund (`bg-slate-900/50`)

### Task 4: Integrate spinner in default layout
- Uppdatera `layouts/default.vue`
- Lägg till LoadingSpinner-komponenten
- Bind till loading store state

### Task 5: Create composable for API calls
- Skapa `composables/useApi.ts`
- Wrapper runt `$fetch` som automatiskt uppdaterar loading state
- Hantera errors och loading automatiskt

### Task 6: Refactor pages to use new composable
- Uppdatera `pages/index.vue` att använda `useApi`
- Behåll bakåtkompatibilitet för andra pages under migration

### Task 7: Add loading indicators for buttons
- Skapa `components/LoadingButton.vue` eller modifiera befintliga knappar
- Visa mini-spinner i knappar under async operationer

## Testing Strategy
- Manuell test: Verifiera spinner visas vid långsamma API-anrop
- Manuell test: Spinner försvinner när data laddats
- Manuell test: Multiple samtidiga requests hanteras korrekt
- Manuell test: Knappar med loading state fungerar

## Acceptance Criteria

- [x] Spinner visas automatiskt vid alla API-anrop
- [x] Spinner har smidig fade-in/fade-out animation
- [x] Bakgrunden är semi-transparent så användaren ser att något händer
- [x] Design matchar dark theme (slate/emerald färgschema)
- [x] Alla sidor (index, products, listings, etc.) visar spinner korrekt
- [x] Inga konstiga hopskutt eller flimmer när spinner visas/döljer

## Files to Create/Modify

### New Files
- `frontend/stores/loading.ts` - Pinia store
- `frontend/components/LoadingSpinner.vue` - Spinner component
- `frontend/composables/useApi.ts` - API wrapper composable

### Modified Files
- `frontend/layouts/default.vue` - Add spinner to layout
- `frontend/pages/index.vue` - Use new composable (example)
- `frontend/app.vue` - Ensure store is initialized

## Technical Notes
- Använd `useFetch` eller `$fetch` beroende på vad projektet redan använder
- Pinia store ska använda `acceptHMRUpdate` för hot reload
- Spinner animation via Tailwind `animate-spin` eller custom CSS
- Consider adding delay (200-300ms) before showing spinner to avoid flash for fast requests

## Out of Scope
- Skeleton loaders (placeholder content while loading)
- Progress bars för långvariga operationer
- Cancel/knapp för att avbryta requests
- Offline indicators

---

## Spec Status: ✅ DONE

**Completed:** 2026-02-03

**Summary:**
All tasks completed. Global loading spinner fully implemented:
- Pinia store (`stores/loading.ts`) med counter-baserad loading state och 200ms delay
- LoadingSpinner komponent (`components/LoadingSpinner.vue`) med semi-transparent overlay och emerald spinner
- Layout integration (`layouts/default.vue`) - spinner visas automatiskt vid alla API-anrop
- API composable (`composables/useApi.ts`) - wrapper runt $fetch som hanterar loading state
- Alla sidor uppdaterade att använda `useApi()` istället för direkt `$fetch`
