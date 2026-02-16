# Relevant Standards

## Configuration Structure

- Use `nuxt.config.ts` for all configuration
- Environment variables for sensitive data (Supabase keys)
- Type-safe config with proper TypeScript types

## Currency Storage

- Display monetary values as SEK with proper formatting
- Store as integers (öre) in API communication
- Format: `{{ value / 100 }} kr`

## LLM Service Functions (for future AI features)

- One function per task
- Action+Entity naming convention

## Tech Stack Standards

- Nuxt 3 directory structure (pages, components, composables)
- Vue 3 Composition API with `<script setup>`
- Tailwind CSS for styling
- Supabase client for auth and database
