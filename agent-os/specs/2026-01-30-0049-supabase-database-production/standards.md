# Standards: Supabase Production Database

## Relevant Standards

### 1. Currency Storage (database/currency-storage.md)
Store all monetary values as integers representing SEK öre (cents).

**Application:**
- The schema already follows this standard (INTEGER for price/cost fields)
- No changes needed
- Go models in `internal/models/models.go` already use `int` for prices

**Example from schema:**
```sql
buy_price INTEGER,
sell_price INTEGER,
```

**Example from Go models:**
```go
BuyPrice  int  // 14900 SEK = 149.00 SEK
SellPrice *int // nullable
```

### 2. Tech Stack (tech-stack.md)
Database should use PostgreSQL.

**Application:**
- Supabase provides managed PostgreSQL
- Perfect match with existing tech stack
- No migration from other database needed

### 3. Email Sending (backend/email-sending.md)
Use `net/smtp` for SMTP sending.

**Note:** This standard is for sending emails, not for database auth. However, Supabase handles email verification and password reset emails automatically. We can reference this if we later add custom email features.

## Security Standards (Implicit)

### SSL/TLS Required
- All database connections must use SSL
- Supabase enforces this by default
- Set `sslmode=require` in connection string

### Credentials Management
- Database credentials must be in config file, never hardcoded
- Use environment variables or secrets management
- Never commit credentials to git

### Authentication
- Email/password authentication required
- Use Supabase Auth for this
- Do not implement custom auth initially
