# Database Provisioning

Run the provisioning script to set up roles, schema, and grants in a new Neon database. Use the owner user for this setup.

```bash
NEON_OWNER_URL="postgres://..." ./provision.sh
```

Passwords will be printed to console.

## Run migrations

```bash
# Set the database URL (postgres:// format with search_path=app)
DATABASE_URL="postgres://atlas:...@host:5432/dbname?sslmode=require&search_path=app"

atlas migrate apply \
  --dir "file://db/migrations" \
  --url "$DATABASE_URL"
```
