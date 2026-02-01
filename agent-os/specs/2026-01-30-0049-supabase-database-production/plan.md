# Plan: Supabase Production Database Setup

## Overview
Set up a production PostgreSQL database on Supabase using the existing schema from `schema_improved.sql`. Enable email/password authentication for secure access.

## Tasks

1. ~~**Save spec documentation**~~ ✅ COMPLETED
   - Document scope, standards, and references

2. ~~**Create Supabase project**~~ ✅ COMPLETED
   - Sign up/log in to Supabase
   - Create new project (begbot)
   - Copy database connection details

3. ~~**Import schema to Supabase**~~ ✅ COMPLETED
   - Use Supabase SQL Editor to run `schema_improved.sql`
   - Verify all tables created successfully
   - Confirm indexes are in place
   - **Result:** 11 tables created and verified

4. ~~**Set up email/password authentication**~~ ✅ COMPLETED
   - Enable email authentication in Supabase
   - Configure email templates (optional)
   - Test authentication flow

5. ~~**Update application config**~~ ✅ COMPLETED
   - Update `config.yaml` with Supabase connection details
   - Ensure SSL is enabled
   - **Important fix:** Switched from `lib/pq` to `pgx` driver for Session Pooler compatibility
   - Connection URL format: `postgresql://user:password@host:port/dbname?sslmode=require`

6. ~~**Test migration and data access**~~ ✅ COMPLETED
   - Run application migration code
   - Insert test data (Product with ID 1 saved successfully)
   - Verify CRUD operations work
   - **Test result:** All CRUD operations working via pgx driver

## Technical Changes Made

### Database Driver Migration
- **Before:** `github.com/lib/pq` (did not work with Supabase Session Pooler)
- **After:** `github.com/jackc/pgx/v5/stdlib` (works correctly)
- **Reason:** Session Pooler returns EOF with lib/pq driver

### Connection String Format
- **Before:** Parameter format `host=X port=Y user=Z...`
- **After:** URL format `postgresql://user:password@host:port/dbname?sslmode=require`
- **Reason:** Better compatibility with modern PostgreSQL drivers

### Session Pooler Configuration
- **Host:** `aws-1-eu-west-1.pooler.supabase.com` (not aws-0)
- **Port:** `5432` (not 6543)
- **User:** `postgres.fxhknzpqhrkpqothjvrx` (project ref appended)
- **Password:** Same as direct connection
- **Reason:** IPv6-only direct connection blocked by local network

## Definition of Done
- ✅ Supabase project is created and accessible
- ✅ Database schema matches `schema_improved.sql` exactly (11 tables verified)
- ✅ Email/password authentication is enabled
- ✅ Application can connect and perform database operations (tested via pgx driver)
- ✅ Connection details are securely configured (not hardcoded)
- ✅ Test data successfully saved and retrieved
