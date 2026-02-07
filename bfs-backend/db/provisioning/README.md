# Database Provisioning

Run the provisioning script to set up roles, schema, and grants in a new Neon database. Use the owner user for this setup.

```bash
NEON_OWNER_URL="postgres://..." ./provision.sh
```

Passwords will be printed to console.

## Run migrations

```bash
# Set environment variables
FLYWAY_URL="jdbc:postgresql://..." FLYWAY_USER=flyway FLYWAY_PASSWORD=...

docker run --rm \
            -v ./db/flyway/migrations:/flyway/sql:ro \         
            -e FLYWAY_URL="${FLYWAY_URL}" \   
            -e FLYWAY_USER="${FLYWAY_USER}" \   
            -e FLYWAY_PASSWORD="${FLYWAY_PASSWORD}" \   
            flyway/flyway:9-alpine \
            -defaultSchema=app \
            -initSql="SET ROLE app_owner" \
            -baselineOnMigrate=true \
            migrate
```