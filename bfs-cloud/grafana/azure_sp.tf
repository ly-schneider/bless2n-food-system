resource "azuread_application" "grafana" {
  display_name = "sp-grafana-cloud-reader"
}

resource "azuread_service_principal" "grafana" {
  client_id = azuread_application.grafana.client_id
}

resource "azuread_service_principal_password" "grafana" {
  service_principal_id = azuread_service_principal.grafana.id
  end_date_relative    = "8760h"
}

data "azurerm_resource_group" "env" {
  for_each = local.env_workspaces
  name     = each.value.resource_group_name
}

resource "azurerm_role_assignment" "grafana_monitoring_reader" {
  for_each             = local.env_workspaces
  scope                = data.azurerm_resource_group.env[each.key].id
  role_definition_name = "Monitoring Reader"
  principal_id         = azuread_service_principal.grafana.object_id
}

resource "azurerm_role_assignment" "grafana_log_analytics_reader" {
  for_each             = local.env_workspaces
  scope                = data.azurerm_resource_group.env[each.key].id
  role_definition_name = "Log Analytics Reader"
  principal_id         = azuread_service_principal.grafana.object_id
}
