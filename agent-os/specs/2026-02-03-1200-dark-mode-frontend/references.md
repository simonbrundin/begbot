# Reference Implementations

## Similar Code Patterns in This Codebase

### 1. Tailwind Configuration
**File:** `/home/simon/repos/begbot/frontend/tailwind.config.js`
- Already defines `primary` color palette (blue tones)
- Used as reference for extending colors

### 2. Main CSS Structure
**File:** `/home/simon/repos/begbot/frontend/assets/css/main.css`
- Defines `@layer base`, `@layer components`, `@layer utilities`
- Pattern to follow for dark mode additions
- Contains component classes: btn, card, input, label, table

### 3. App Entry Point
**File:** `/home/simon/repos/begbot/frontend/app.vue`
- Uses NuxtLayout wrapper
- Theme class should be applied here or in nuxt.config

### 4. Nuxt Config
**File:** `/home/simon/repos/begbot/frontend/nuxt.config.ts`
- Configures Tailwind via `@nuxtjs/tailwindcss`
- CSS entry point: `~/assets/css/main.css`

## External References (Not Included)

Tailwind CSS Dark Mode Documentation:
- https://tailwindcss.com/docs/dark-mode
- Class-based dark mode pattern
- Customizing colors for dark mode
