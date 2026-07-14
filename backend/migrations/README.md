# Database Migrations

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations in production.

## Installation

### macOS
```bash
brew install golang-migrate
```

### Linux
```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

### Windows
```bash
scoop install migrate
```

Or download from [releases](https://github.com/golang-migrate/migrate/releases).

## Creating Migrations

```bash
# Create a new migration
migrate create -ext sql -dir migrations -seq <migration_name>

# Example: Create users table
migrate create -ext sql -dir migrations -seq create_users_table
```

This creates two files:
- `000001_create_users_table.up.sql` - Apply migration
- `000001_create_users_table.down.sql` - Rollback migration

## Running Migrations

```bash
# Set database URL
export DATABASE_URL="postgres://user:password@localhost:5432/fiber_api?sslmode=disable"

# Apply all migrations
migrate -database "$DATABASE_URL" -path migrations up

# Apply N migrations
migrate -database "$DATABASE_URL" -path migrations up N

# Rollback last migration
migrate -database "$DATABASE_URL" -path migrations down 1

# Rollback all migrations
migrate -database "$DATABASE_URL" -path migrations down

# Check current version
migrate -database "$DATABASE_URL" -path migrations version

# Force version (fix dirty state)
migrate -database "$DATABASE_URL" -path migrations force VERSION
```

## Example Migration Files

### 000001_create_users_table.up.sql
```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

### 000001_create_users_table.down.sql
```sql
DROP TABLE IF EXISTS users;
```

## Best Practices

1. **Always test migrations locally** before applying to production
2. **Never modify** existing migration files after they've been applied
3. **Use transactions** when possible (PostgreSQL supports transactional DDL)
4. **Backup database** before running migrations in production
5. **Keep migrations small** and focused on single changes

## CI/CD Integration

```yaml
# Example GitHub Actions step
- name: Run migrations
  env:
    DATABASE_URL: ${{ secrets.DATABASE_URL }}
  run: |
    migrate -database "$DATABASE_URL" -path migrations up
```
