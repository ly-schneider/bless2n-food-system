#!/bin/bash
set -e

if [ -z "$NEON_OWNER_URL" ]; then
  echo "Error: NEON_OWNER_URL environment variable is not set"
  exit 1
fi

cd "$(dirname "$0")"

echo "Running provisioning scripts..."
psql "$NEON_OWNER_URL" \
  -f 01_roles.sql \
  -f 02_schema.sql \
  -f 03_grants_existing.sql \
  -f 04_default_privileges.sql \
  -f 05_role_settings.sql \
  -f 06_verify.sql

echo "Setting passwords..."

ATLAS_PW=$(openssl rand -hex 32)
APP_BACKEND_PW=$(openssl rand -hex 32)
APP_ADMIN_PW=$(openssl rand -hex 32)

psql "$NEON_OWNER_URL" <<EOF
ALTER USER atlas       WITH PASSWORD '$ATLAS_PW';
ALTER USER app_backend WITH PASSWORD '$APP_BACKEND_PW';
ALTER USER app_admin   WITH PASSWORD '$APP_ADMIN_PW';
EOF

echo ""
echo "atlas:       $ATLAS_PW"
echo "app_backend: $APP_BACKEND_PW"
echo "app_admin:   $APP_ADMIN_PW"
