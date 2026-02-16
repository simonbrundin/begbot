# Relevant Standards for Loading Spinner

## Frontend Standards

### Swedish Text Standard
**Reference**: `agent-os/standards/frontend/swedish-text.md`

**Requirements**:
- All text i UI ska vara på svenska
- Använd naturlig svensk formulering

**Applicability**: 
- `aria-label` och eventuella text-meddelanden i spinner
- Exempel: `aria-label="Laddar..."` istället för `"Loading..."`

### Tech Stack Standard
**Reference**: `agent-os/standards/tech-stack.md`

**Requirements**:
- React 18 with TypeScript
- Tailwind CSS v4 for styling
- Vite for build tooling

**Applicability**:
- Component ska skrivas i Vue (not React - project actually uses Nuxt/Vue)
- Tailwind v6 (per package.json) för styling
- TypeScript för type safety

## Global Standards

### Configuration Structure
**Reference**: `agent-os/standards/global/configuration-structure.md`

**Requirements**:
- Nested config structs
- time.Duration timeouts

**Applicability**:
- Loading store config (delay threshold, animation duration)

## Applying Standards

### In This Feature
1. **Swedish Text**: Alla aria-labels och eventuella text-meddelanden på svenska
2. **TypeScript**: Full type coverage i store och composable
3. **Tailwind CSS**: Inga custom CSS, all styling via Tailwind utilities
4. **Component Structure**: Följ befintliga Vue 3 patterns i projektet

### Standards to Follow
- [x] Swedish text in UI
- [x] TypeScript types
- [x] Tailwind CSS styling
- [x] Vue 3 Composition API
- [x] Pinia for state management
