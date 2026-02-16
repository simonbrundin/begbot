# Scraping Cancel Button - Shape

## Scope
**In scope:**
- Avbryt-knapp på /scraping-sidan
- Backend cancellation med context/channel
- Spara partial data vid avbryt
- Ny "cancelled" status

**Out of scope:**
- Job history/persistence över server restart
- Återuppta avbrutna jobb
- Cancel för flera samtidiga jobb (hanteras redan av designen)

## Key Decisions

### 1. Cancellation Mechanism
**Val:** Channel-based cancellation i FetchJob struct  
**Motivering:** Enkel att implementera, thread-safe, ingen external dependency  
**Alternativ:** Context.Context - mer idiomatiskt för Go men kräver större refaktorering av BotService  

### 2. Partial Data Handling
**Val:** Spara all data som redan scrapats  
**Motivering:** Användaren ville spara insamlad data  
**Implikation:** BotService måste hantera "clean exit" och inte rulla tillbaka transaktioner

### 3. UI Placement
**Val:** Knapp bredvid progress-indikatorn i status-kortet  
**Motivering:** Tydlig koppling till pågående jobb, lättillgänglig  
**Design:** Röd knapp med text "Avbryt", disabled state under anrop

### 4. Status Flow
```
pending → cancelled (om avbruten innan start)
running → cancelled (om avbruten under körning)
```

## Technical Constraints
- Använd existerande mutex-mönster från JobService
- Följ Go-konventioner för channel-hantering
- Svenska texter enligt standards
- Behåll existerande API-struktur

## Risks & Mitigations
| Risk | Sannolikhet | Påverkan | Mitigering |
|------|-------------|----------|------------|
| Goroutine leak vid cancel | Låg | Medium | Se till att all goroutines kollar cancel-channel |
| Race condition i status | Låg | Medium | Använd mutex korrekt i JobService |
| UI visar fel status | Låg | Låg | Tester för alla status-övergångar |

## References
- Go Concurrency Patterns: https://go.dev/blog/pipelines
- Existing job flow: `cmd/api/main.go:539-570`
