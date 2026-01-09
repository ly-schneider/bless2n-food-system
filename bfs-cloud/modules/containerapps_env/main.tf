# Container Apps Environment - external without VNet injection (cost-optimized)
# This creates a public-facing environment that uses Azure's shared infrastructure
# instead of dedicated networking resources, eliminating MC_* resource group costs.
resource "azurerm_container_app_environment" "this" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name

  log_analytics_workspace_id = var.log_analytics_workspace_id

  tags = var.tags
}
