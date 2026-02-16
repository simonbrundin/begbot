# Database Frontend Integration Plan

## Problem Statement
Data from database is not displaying in the frontend when running dev servers via `simon dev`.

## Root Cause Found

**Port mismatch**: 
- Backend runs on port **8081** (`dev.nu` line 27)
- Frontend defaults to `http://localhost:8080` when `API_BASE_URL` not set (`nuxt.config.ts` line 27)
- `dev.nu` does NOT set `API_BASE_URL` environment variable

**Verification**:
```
curl http://localhost:8081/api/health → {"status":"ok"}
curl http://localhost:8080/api/health → connection refused
```

## Solution

Update `dev.nu` to set `API_BASE_URL` when starting frontend:

```bash
cd /home/simon/repos/begbot/frontend
API_BASE_URL="http://localhost:8081" npm run dev
```

## Tasks

1. ✅ Save spec documentation
2. ✅ Identify root cause (port mismatch)
3. ✅ Fix `dev.nu` to set `API_BASE_URL`
4. ✅ Verify fix works - run `simon dev` and check data displays

**Additional Fix Required:**
During verification, discovered and fixed NULL handling issue:
- **Problem**: Products API failed with "converting NULL to string is unsupported" for `model_variant` column
- **Root Cause**: `Product.ModelVariant` field was `string` type but database allows NULL values
- **Solution**: Changed `ModelVariant` from `string` to `*string` in `internal/models/models.go:12`
- **Also Fixed**: Updated `GetOrCreateProduct` function signature in `internal/db/postgres.go:372` to accept `*string`
- **Result**: All APIs now working correctly:
  - ✅ `/api/products` - Returns 18 products
  - ✅ `/api/listings` - Returns listings data
  - ✅ `/api/inventory` - Returns inventory items
  - ✅ Frontend can now display database data

## Implementation Details

**File changed**: `dev.nu` (line 93)

**Before**:
```bash
^bash -c "npm run dev > /dev/null 2>&1 &"
```

**After**:
```bash
^bash -c $"API_BASE_URL='http://localhost:8081' npm run dev > /dev/null 2>&1 &"
```

## Verification Steps

1. Run `simon dev` and select "all"
2. Open http://localhost:3000
3. Navigate to /listings, /products, or /
4. Verify data from Supabase is now visible

## Spec Folder
`agent-os/specs/2026-02-03-1100-database-frontend-integration/`

---

## Spec Status: ✅ DONE

**Completed:** 2026-02-03

**Summary:**
- Fixed port mismatch between backend (8081) and frontend (8080)
- Fixed NULL handling for `model_variant` column in products
- All API endpoints working correctly
- Frontend can now successfully display data from Supabase database
