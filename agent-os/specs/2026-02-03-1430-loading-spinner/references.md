# Reference Implementations

## Existing Code Patterns

### API Fetching Pattern
**File**: `frontend/pages/index.vue:219-233`

```typescript
const fetchData = async () => {
  loading.value = true
  try {
    const [itemsRes, productsRes] = await Promise.all([
      $fetch<TradedItem[]>(`${config.public.apiBase}/api/inventory`),
      $fetch<Product[]>(`${config.public.apiBase}/api/products`)
    ])
    items.value = itemsRes
    products.value = productsRes
  } catch (e) {
    console.error('Failed to fetch data:', e)
  } finally {
    loading.value = false
  }
}
```

**Key Insights**:
- Använder `$fetch` direkt
- Manuell loading state hantering per page
- Error handling med console.error

### Dark Theme Colors
**File**: `frontend/layouts/default.vue:2`

```vue
<div class="min-h-screen bg-slate-900 text-slate-100">
```

**Key Insights**:
- Bakgrund: `bg-slate-900`
- Text: `text-slate-100`
- Sidonav: `bg-slate-800` med `border-slate-700`

### Emerald Accent Color
**File**: `frontend/layouts/default.vue:7`

```vue
<h1 class="text-xl font-bold text-emerald-400">Begbot</h1>
```

**Key Insights**:
- Brand color: `emerald-400`
- Använd för spinner animation eller accent

### Layout Structure
**File**: `frontend/layouts/default.vue`

**Key Insights**:
- Fixed sidebar: `w-64`, `left-0`, `top-0`
- Main content: `ml-64` (offset för sidebar)
- Spinner bör vara `fixed` och täcka hela viewport

## Component Patterns

### Existing Button Component Pattern
**File**: `frontend/pages/index.vue:5-7`

```vue
<button @click="showAddModal = true" class="btn btn-primary">
  Lägg till
</button>
```

**Key Insights**:
- Använder utility classes (`btn`, `btn-primary`)
- Tailwind-based styling
- Event handlers direkt på element

### Modal Pattern
**File**: `frontend/pages/index.vue:91-92`

```vue
<div v-if="showAddModal || editingItem" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
```

**Key Insights**:
- `fixed inset-0` för full viewport coverage
- `z-50` för högsta z-index
- `bg-black/70` för semi-transparent overlay

## API Structure

### Base URL Configuration
**File**: `frontend/pages/index.vue:175`

```typescript
const config = useRuntimeConfig()
// Används som: ${config.public.apiBase}/api/inventory
```

**Key Insights**:
- `useRuntimeConfig()` för API base URL
- Pattern: `${config.public.apiBase}/api/{endpoint}`

## State Management References

### Current Pattern
**File**: `frontend/pages/index.vue:179-181`

```typescript
const loading = ref(true)
const showAddModal = ref(false)
const editingItem = ref<TradedItem | null>(null)
```

**Key Insights**:
- Använder `ref()` för reactive state
- TypeScript generics för typed refs
- Initial values sätts direkt

## Project Structure

### Relevant Directories
```
frontend/
├── components/          # Place LoadingSpinner.vue here
├── composables/          # Place useApi.ts here
├── layouts/             # Modify default.vue
├── pages/               # Pages to integrate
├── stores/              # Create loading.ts here
└── types/               # Type definitions
```

### Files to Read for Context
- `frontend/nuxt.config.ts` - Pinia och runtime config
- `frontend/tailwind.config.js` - Custom colors/animations
- `frontend/types/database.ts` - Type definitions
