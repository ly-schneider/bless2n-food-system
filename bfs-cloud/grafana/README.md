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

### 4. Terraform Cloud workspace

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

Defaults in `variables.tf` cover `grafana_stack_url`, Sentry org slug, and the remote state workspace names — override as needed.

### 5. Grant remote state access

In each env workspace (`staging`, `production`) → **Settings → Remote state sharing** → allow this workspace.

### 6. Apply

```
terraform init
terraform plan
terraform apply
```

Expected resources:

- 2 `grafana_data_source` (Azure Monitor, one per env)
- 0–1 `grafana_data_source` (Sentry)
- 1 `grafana_folder`
- 1 `grafana_dashboard`

## Updating the dashboard

**Small tweaks:** edit `dashboards/bfs-overview.json` directly and apply. `overwrite = true` lets the provider replace UI edits.

**Larger changes:**

1. Edit in the Grafana UI
2. **Share → Export → "Export for sharing externally"**
3. Paste the JSON into `dashboards/bfs-overview.json`
4. Confirm datasource refs use `"uid": "${datasource}"` and `"uid": "sentry"`
5. Commit and apply

## Dashboard variables

| Variable                           | Used by              | Behavior                                            |
| ---------------------------------- | -------------------- | --------------------------------------------------- |
| `$datasource`                      | Azure Monitor panels | Dropdown listing both env data sources              |
| `$env`                             | Sentry panels        | Staging / production filter                         |
| `$subscription` / `$resourceGroup` | metric panels        | Auto-populated from the selected data source        |
| `$app`                             | metric panels        | Multi-select over Container Apps in the selected RG |

Switch `$datasource` to toggle the whole dashboard between envs. For Sentry panels to follow, also switch `$env`.

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
