# Grafana Cloud — IaC

Terraform that manages the Grafana Cloud stack at https://leysservices.grafana.net, its Azure Monitor + Sentry data sources, and the consolidated observability dashboard for both staging and production.

## Layout

```
grafana/
├── versions.tf          # providers + TF Cloud backend (workspace: bfs-grafana)
├── variables.tf         # inputs (see "Variables" below)
├── main.tf              # remote state, folder, data sources, dashboard
├── azure_sp.tf          # sp-grafana-cloud-reader + RG role assignments
├── outputs.tf
└── dashboards/
    └── bfs-overview.json
```

## One-time manual setup

### 1. Grafana Cloud access policy + token

1. Log in to https://leysservices.grafana.net.
2. **Administration → Users and access → Cloud access policies → Create access policy**.
3. Name: `terraform-admin`. Scopes:
   - `stacks:read`
   - `dashboards:write`
   - `folders:write`
   - `datasources:write`
   - `alerts:write`
4. Add a token under the policy. Copy it — this is `grafana_cloud_token`.

### 2. Sentry internal integration token (optional, wires the Errors row)

1. https://sentry.io → **Settings → Developer Settings → New Internal Integration**.
2. Name: `grafana-cloud-reader`. Permissions:
   - Organization: Read
   - Project: Read
   - Event: Read
   - Issue & Event: Read
3. Copy the token — this is `sentry_auth_token`.

### 3. Create the TF Cloud workspace `bfs-grafana`

- Organization: `leys-services`
- VCS-backed on this repo; **working directory**: `bfs-cloud/grafana`
- Execution mode: Remote
- Variables (set all as **Terraform variables**, not env):

| Name                    | Kind   | Sensitive | Notes                                           |
| ----------------------- | ------ | --------- | ----------------------------------------------- |
| `grafana_cloud_token`   | string | ✅        | From step 1                                     |
| `sentry_auth_token`     | string | ✅        | From step 2 (or empty string to skip)           |
| `azure_tenant_id`       | string | —         | Same tenant as `bfs-staging` / `bfs-production` |
| `azure_subscription_id` | string | —         | Subscription containing both env RGs            |

Defaults in `variables.tf` cover `grafana_stack_url`, Sentry org slug, and the remote state workspace names — override only if they diverge.

### 4. Grant the workspace access to remote state

In the `bfs-staging` and `bfs-production` workspaces → **Settings → General → Remote state sharing** → add `bfs-grafana` to the allow list. This lets it read `log_analytics_workspace_id` and `resource_group_name` outputs.

### 5. First apply

```
cd bfs-cloud/grafana
terraform init
terraform plan
terraform apply
```

Expected resources on first apply:

- 1 `azuread_application` (`sp-grafana-cloud-reader`)
- 1 `azuread_service_principal`
- 1 `azuread_service_principal_password` (1y lifetime)
- 4 `azurerm_role_assignment` (Monitoring Reader + Log Analytics Reader × 2 envs)
- 2 `grafana_data_source` (Azure Monitor staging + production)
- 0–1 `grafana_data_source` (Sentry, if token provided)
- 1 `grafana_folder`
- 1 `grafana_dashboard`

## Updating the dashboard

Two workflows depending on how far you're iterating:

**Small tweaks (query, panel title, threshold):** edit `dashboards/bfs-overview.json` directly, `terraform apply`. The `overwrite = true` flag means Grafana accepts the new version even if the UI was edited in the meantime.

**Larger changes (add panels, restructure):**

1. In the Grafana UI, edit the dashboard.
2. **Share → Export → "Export for sharing externally"** → copy JSON.
3. Paste into `dashboards/bfs-overview.json`.
4. Sanity-check: datasource refs should use `"uid": "${datasource}"` and `"uid": "sentry"` (the exporter templates them automatically).
5. Commit + `terraform apply`.

## Dashboard variables

| Variable                           | Used by                  | Behavior                                                      |
| ---------------------------------- | ------------------------ | ------------------------------------------------------------- |
| `$datasource`                      | all Azure Monitor panels | Dropdown listing both Azure Monitor data sources; selects env |
| `$env`                             | Sentry panels            | `staging` / `production`, filters Sentry environments         |
| `$subscription` / `$resourceGroup` | metric panels            | Auto-populated from the selected data source                  |
| `$app`                             | metric panels            | Multi-select over Container Apps in the selected RG           |

Toggling `$datasource` between `Azure Monitor - staging` and `Azure Monitor - production` switches the entire dashboard. For Sentry to follow, also switch `$env`.

## Known gaps / follow-ups

- **Latency panel** parses a `latency_ms` JSON field from `bfs-backend` Zap logs. If that field name differs, update the KQL in panel id `6`. If the backend doesn't emit request latency yet, the panel stays empty — add a Zap middleware field (`zap.Duration("latency_ms", ...)`).
- **Status codes panel** similarly parses a `status` field from JSON logs. Same caveat.
- **Alert rules** are not yet defined — add `grafana_rule_group` resources once thresholds are calibrated.
- **Synthetic monitoring** for true external SLA (public ordering URL) can be added via `grafana_synthetic_monitor_check`.

## Secret rotation

The Azure SP password expires 1 year after creation. Rotate:

```
terraform apply -replace=azuread_service_principal_password.grafana
```

The Azure Monitor data sources will re-apply with the new secret automatically.
