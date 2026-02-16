# Reference Implementations

## Backend Database Functions

- `internal/db/postgres.go` - All database CRUD operations
- `internal/models/models.go` - Data models (TradedItem, Listing, Product, etc.)

## Existing Backend Patterns

- `internal/services/` - Service layer pattern for business logic
- `internal/config/config.go` - Configuration loading pattern

## Supabase

- Use `@nuxtjs/supabase` module
- Auth flow: https://nuxt.com/modules/supabase

## Vue/Nuxt Best Practices

- Vue 3 Composition API: `<script setup lang="ts">`
- Nuxt 3 auto-imports for components and composables
- Tailwind CSS utility classes
- TypeScript for all files
