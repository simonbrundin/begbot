#!/bin/bash
set -e

DB_NAME="${DB_NAME:-begbot}"
DB_USER="${DB_USER:-begbot}"
DB_PASSWORD="${DB_PASSWORD:-begbot}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"

echo "Creating database and user..."

psql -h "$DB_HOST" -p "$DB_PORT" -U postgres <<EOF
-- Create user if not exists
DO
\$do\$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = '$DB_USER') THEN
      CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
   END IF;
END
\$do\$;

-- Create database
CREATE DATABASE $DB_NAME OWNER $DB_USER;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;

-- Connect to the database and grant schema permissions
\c $DB_NAME
GRANT ALL ON SCHEMA public TO $DB_USER;
EOF

echo "Database setup complete!"
echo "Connection string: postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME"
