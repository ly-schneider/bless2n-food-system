#!/bin/bash
set -e

if [ -z "$DATABASE_URL" ]; then
  echo "Error: DATABASE_URL environment variable is not set"
  exit 1
fi

cd "$(dirname "$0")"

echo "Running provisioning scripts..."
psql "$DATABASE_URL" \
  -f 01_roles.sql \
  -f 02_grants_existing.sql \
  -f 03_default_privileges.sql \
  -f 04_role_settings.sql \
  -f 05_verify.sql

echo "Provisioning complete."
