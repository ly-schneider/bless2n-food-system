# Database Provisioning

## Local Development

Roles are created automatically on first `docker compose up` via `docker-init.sql` (mounted into the Postgres init directory). Hardcoded dev passwords are set for `atlas`, `app_backend`, and `app_admin`.

```bash
just up          # roles auto-provisioned on fresh volume
just migrate     # apply migrations
just seed        # seed dev data
just dev         # start backend
```

To re-provision from scratch: `just down --volumes && just up`.

## New Remote Environment

1. Run `provision.sh` (or let CD do it automatically):
   ```bash
   DATABASE_URL="postgres://owner:...@host/db" ./provision.sh
   ```
2. Set passwords (manual, once per environment):
   ```bash
   DATABASE_URL="postgres://owner:...@host/db" ./set-passwords.sh
   ```
3. Store the printed passwords in your secret manager.

## Existing Environment (CD)

`provision.sh` runs automatically before every migration in the CD pipeline. It is fully idempotent — roles that already exist are skipped, grants are re-applied as no-ops.

## Scripts

| Script | Purpose | When to run |
|---|---|---|
| `provision.sh` | Roles, grants, default privileges | Automatic in CD; manual for new environments |
| `set-passwords.sh` | Generate and set random passwords | Manual, once per environment |
| `docker-init.sql` | Combined init for local Docker | Automatic on `docker compose up` (fresh volume) |
