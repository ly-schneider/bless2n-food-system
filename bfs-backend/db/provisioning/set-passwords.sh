#!/bin/bash
set -e

if [ -z "$DATABASE_URL" ]; then
  echo "Error: DATABASE_URL environment variable is not set"
  exit 1
fi

ATLAS_PW=$(openssl rand -hex 32)
APP_BACKEND_PW=$(openssl rand -hex 32)
APP_ADMIN_PW=$(openssl rand -hex 32)

psql "$DATABASE_URL" <<EOF
ALTER USER atlas       WITH PASSWORD '$ATLAS_PW';
ALTER USER app_backend WITH PASSWORD '$APP_BACKEND_PW';
ALTER USER app_admin   WITH PASSWORD '$APP_ADMIN_PW';
EOF

echo ""
echo "atlas:       $ATLAS_PW"
echo "app_backend: $APP_BACKEND_PW"
echo "app_admin:   $APP_ADMIN_PW"
