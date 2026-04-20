data "terraform_remote_state" "staging" {
  backend = "remote"
  config = {
    organization = "leys-services"
    workspaces = {
      name = var.staging_remote_state_workspace
    }
  }
}

data "terraform_remote_state" "production" {
  backend = "remote"
  config = {
    organization = "leys-services"
    workspaces = {
      name = var.production_remote_state_workspace
    }
  }
}

locals {
  env_law_ids = {
    staging    = data.terraform_remote_state.staging.outputs.log_analytics_workspace_id
    production = data.terraform_remote_state.production.outputs.log_analytics_workspace_id
  }

  sentry_enabled = length(var.sentry_auth_token) > 0
}

resource "grafana_folder" "bfs" {
  title = "Bless2n Food System"
}

resource "grafana_data_source" "azure_monitor" {
  for_each = local.env_law_ids

  type = "grafana-azure-monitor-datasource"
  name = "Azure Monitor - ${each.key}"
  uid  = "azmon-${each.key}"

  json_data_encoded = jsonencode({
    azureAuthType                = "clientsecret"
    cloudName                    = "azuremonitor"
    tenantId                     = var.azure_tenant_id
    clientId                     = var.grafana_azure_client_id
    subscriptionId               = var.azure_subscription_id
    logAnalyticsDefaultWorkspace = each.value
  })

  secure_json_data_encoded = jsonencode({
    clientSecret = var.grafana_azure_client_secret
  })
}

resource "grafana_data_source" "sentry" {
  count = local.sentry_enabled ? 1 : 0

  type = "grafana-sentry-datasource"
  name = "Sentry"
  uid  = "sentry"

  json_data_encoded = jsonencode({
    url     = "https://sentry.io"
    orgSlug = var.sentry_organization_slug
  })

  secure_json_data_encoded = jsonencode({
    authToken = var.sentry_auth_token
  })
}

resource "grafana_dashboard" "overview" {
  folder      = grafana_folder.bfs.uid
  config_json = file("${path.module}/dashboards/bfs-overview.json")
  overwrite   = true

  depends_on = [
    grafana_data_source.azure_monitor,
    grafana_data_source.sentry,
  ]
}
