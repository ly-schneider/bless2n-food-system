resource "azurerm_container_app_environment" "this" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name

  infrastructure_subnet_id       = var.subnet_id
  internal_load_balancer_enabled = false

  log_analytics_workspace_id = var.logs_destination == "log-analytics" ? var.log_analytics_workspace_id : null

  tags = var.tags
}
