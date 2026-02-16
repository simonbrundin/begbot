# Loading Spinner - Shaping Notes

## Problem
När användaren navigerar mellan sidor eller utför API-anrop finns det ingen visuell feedback om att något händer. Detta leder till:
- Användaren vet inte om appen är responsiv
- Dubbelklick på knappar kan ske
- Känsla av att appen är "seg" även om API:et är snabbt

## Solution
Global loading spinner som:
1. Visas automatiskt vid alla API-anrop
2. Har smidig animation (inte för distraherande)
3. Matchar appens dark theme
4. Är semi-transparent så användaren fortfarande ser innehållet

## Key Decisions

### Why Global State?
Lokala `loading` refs i varje komponent fungerar men:
- Kräver duplicerad kod
- Svårt att hantera när flera requests körs samtidigt
- Inkonsekvent beteende mellan sidor

Global state via Pinia ger bättre kontroll och återanvändbarhet.

### Why Overlay Approach?
Alternativ: Skeleton loaders, inline spinners, progress bars

Valde overlay för att:
- Tydligast signal till användaren
- Fungerar med alla typer av innehåll (tabeller, formulär, etc.)
- Enklare att implementera globalt
- Matchar moderna webbapp-mönster

### Why Semi-Transparent?
Fullt blockerande overlay känns "tungt". Semi-transparent gör att:
- Användaren ser att något händer
- Appen känns mer responsiv
- Mindre disruptive för användarupplevelsen

## Constraints
- Måste fungera med existerande `$fetch` API-anrop
- Ska inte kräva stora refaktoreringar av befintliga pages
- Måste matcha dark theme (slate-900 bakgrund, emerald accent)
- Ska fungera med Tailwind CSS v6

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Spinner blinkar för snabba requests | Lägg till 200ms delay innan visning |
| Flera requests samtidigt hanteras fel | Använd counter, inte boolean |
| Z-index konflikter | Använd z-50 (högst) i layout |
| Accessibility issues | Aria-labels och reduced-motion support |

## Success Metrics
- Spinner visas vid varje API-anrop
- Animation är smidig (60fps)
- Inga konstiga flimmer eller hopskutt
- Design är konsistent med resten av appen
