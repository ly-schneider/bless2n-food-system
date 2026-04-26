# Grafana Cloud — IaC

Terraform that manages a Grafana Cloud stack, its Azure Monitor + Sentry data sources, and a consolidated observability dashboard for staging and production environments.

## Layout

```
grafana/
├── versions.tf          # providers + Terraform Cloud backend
├── variables.tf         # inputs
├── main.tf              # remote state, folder, data sources, dashboard
├── outputs.tf
└── dashboards/
    └── bfs-overview.json
```

## Prerequisites

- A Grafana Cloud stack (free tier is enough)
- An Azure subscription with one resource group per environment, each containing:
  - A Log Analytics workspace
  - Container Apps emitting diagnostic logs/metrics to that workspace
- Existing Terraform Cloud workspaces for each env (`staging`, `production`) that publish `log_analytics_workspace_id` + `resource_group_name` outputs
- A Sentry org (optional — data source can be skipped)

## One-time setup

### 1. Grafana service account + token

Dashboards, folders, and data sources live inside the stack — manage them with a stack service account (not a cloud access policy).

1. In the Grafana stack UI → **Administration → Users and access → Service accounts → Add service account**
2. Role: `Admin`. Create.
3. **Add service account token** → copy it.

### 2. Azure service principal

Create the SP once, outside Terraform (the TF execution identity usually lacks AAD write rights):

```bash
az ad sp create-for-rbac \
  --name sp-grafana-cloud-reader \
  --role "Monitoring Reader" \
  --scopes /subscriptions/<SUB_ID>/resourceGroups/<STAGING_RG> \
           /subscriptions/<SUB_ID>/resourceGroups/<PRODUCTION_RG>

OBJECT_ID=$(az ad sp list --display-name sp-grafana-cloud-reader --query '[0].id' -o tsv)
for RG in <STAGING_RG> <PRODUCTION_RG>; do
  az role assignment create \
    --assignee-object-id "$OBJECT_ID" \
    --assignee-principal-type ServicePrincipal \
    --role "Log Analytics Reader" \
    --scope "/subscriptions/<SUB_ID>/resourceGroups/$RG"
done
```

Keep `appId`, `password`, and `tenant` from the first command's output.

If no password is printed:

```bash
az ad app credential reset \
  --id $(az ad app list --display-name sp-grafana-cloud-reader --query '[0].appId' -o tsv) \
  --years 1
```

### 3. Sentry token (optional)

1. Sentry → **Settings → Developer Settings → New Internal Integration**
2. Permissions: Organization Read, Project Read, Issue & Event Read
3. Copy the token (leave empty to skip Sentry wiring).

### 4. Neon read-only role (optional, unlocks business metrics panels)

The business rows on the overview dashboard (Revenue, Products, Order health) query Postgres directly. Create a `grafana_reader` role on **each** Neon branch (`staging`, `production`), using the same password.

Run the role creation once (Neon roles are cluster-wide):

```sql
CREATE ROLE grafana_reader WITH LOGIN PASSWORD '<generate-and-store>';
```

Then connect to the `bless2n_food_system` database on each branch and run the GRANTs there. The `ALTER DEFAULT PRIVILEGES` statement should be run as the role that owns/creates the tables (the migration role) so that `grafana_reader` automatically picks up future tables:

```sql
GRANT CONNECT ON DATABASE bless2n_food_system TO grafana_reader;
GRANT USAGE ON SCHEMA public TO grafana_reader;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO grafana_reader;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO grafana_reader;
```

Grab the connection host for each branch from the Neon console (e.g. `ep-xxxx-staging.eu-central-1.aws.neon.tech`). Leave `neon_pg_host_*` empty to skip the Postgres data source — the business panels will show "No data" but the rest of the dashboard keeps working.

### 5. Terraform Cloud workspace

Create a workspace backed by this repo, working directory `bfs-cloud/grafana`.

Terraform variables:

| Name                          | Sensitive | Source                                |
| ----------------------------- | --------- | ------------------------------------- |
| `grafana_cloud_token`         | ✅        | Step 1                                |
| `grafana_azure_client_id`     | —         | Step 2 (`appId`)                      |
| `grafana_azure_client_secret` | ✅        | Step 2 (`password`)                   |
| `azure_tenant_id`             | —         | Step 2 (`tenant`)                     |
| `azure_subscription_id`       | —         | Your subscription holding the env RGs |
| `sentry_auth_token`           | ✅        | Step 3 (or `""` to skip)              |
| `neon_pg_host_staging`        | —         | Step 4 (or `""` to skip Postgres)     |
| `neon_pg_host_production`     | —         | Step 4 (or `""` to skip Postgres)     |
| `neon_grafana_password`       | ✅        | Step 4 (or `""` to skip Postgres)     |

Defaults in `variables.tf` cover `grafana_stack_url`, Sentry org slug, the Neon database name (`bless2n_food_system`), and the remote state workspace names — override as needed.

### 6. Grant remote state access

In each env workspace (`staging`, `production`) → **Settings → Remote state sharing** → allow this workspace.

### 7. Apply

```
terraform init
terraform plan
terraform apply
```

Expected resources:

- 2 `grafana_data_source` (Azure Monitor, one per env)
- 0–1 `grafana_data_source` (Sentry)
- 0–2 `grafana_data_source` (Postgres, one per env with a host set)
- 1 `grafana_folder`
- 1 `grafana_dashboard`

## Updating the dashboard

**Small tweaks:** edit `dashboards/bfs-overview.json` directly and apply. `overwrite = true` lets the provider replace UI edits.

**Larger changes:**

1. Edit in the Grafana UI
2. **Share → Export → "Export for sharing externally"**
3. Paste the JSON into `dashboards/bfs-overview.json`
4. Confirm datasource refs use `"uid": "azmon-${env}"`, `"uid": "pg-${env}"`, and `"uid": "sentry"`
5. Commit and apply

## Dashboard variables

| Variable        | Used by                     | Behavior                                                                       |
| --------------- | --------------------------- | ------------------------------------------------------------------------------ |
| `$env`          | all panels                  | Single switch. Resolves data sources as `azmon-${env}`, `pg-${env}`, `sentry`. |
| `$subscription` | Azure Monitor panels        | Hidden. Azure subscription holding both env resource groups.                   |
| `$app`          | Azure Monitor metric panels | Query-populated over container apps in `bfs-${env}-rg`.                        |

One switch (`$env`) flips Azure Monitor, Postgres, and Sentry data together — no more desynced views.

## Known gaps

- **Latency / status-code panels** parse JSON fields (`latency_ms`, `status`) from structured backend logs via `ContainerAppConsoleLogs_CL`. Adjust the KQL if your log shape differs.
- **Alert rules** are not yet defined.
- **Synthetic monitoring** for external SLA is not yet wired up.

## Secret rotation

```bash
az ad app credential reset \
  --id $(az ad app list --display-name sp-grafana-cloud-reader --query '[0].appId' -o tsv) \
  --years 1
```

Paste the new `password` into `grafana_azure_client_secret` and queue a new apply.
