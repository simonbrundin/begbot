# Database Migration Strategy

## Overview

This directory contains database migrations for the begbot project. Migrations are version-controlled SQL files that transform the database schema.

## Naming Convention

- Format: `XXX_description.sql` where XXX is a 3-digit sequence number
- Examples: `001_initial_schema.sql`, `002_add_user_table.sql`

## Applying Migrations

### Supabase CLI (Recommended)

```bash
supabase migration up
supabase db push
```

### Manual (Supabase SQL Editor)

1. Open Supabase SQL Editor
2. Select the appropriate database connection
3. Run migrations in chronological order

## Migration Best Practices

1. **Always be backward compatible** - Don't drop columns or tables without migration strategy
2. **Idempotent migrations** - Can be run multiple times without side effects
3. **Test migrations** - Apply in development before production
4. **Document changes** - Add comments explaining the purpose of each migration

## Current State

- `001_initial_schema.sql` - Initial schema from `schema_improved.sql`
  - Creates 11 tables
  - Sets up indexes for performance
  - Configures RLS policies

## Rollback Strategy

Supabase migrations don't support automatic rollback. To rollback:

1. Create a new migration with reversing SQL
2. Apply the rollback migration
3. Document the rollback in this README

Example rollback migration:
```sql
-- Rollback: Remove user_preferences table
DROP TABLE IF EXISTS user_preferences;
```
