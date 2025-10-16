resource "azurerm_monitor_diagnostic_setting" "this" {
  name                           = var.name
  target_resource_id             = var.target_resource_id
  log_analytics_workspace_id     = var.log_analytics_workspace_id
  log_analytics_destination_type = "Dedicated"

  dynamic "enabled_log" {
    for_each = var.categories
    content {
      category = enabled_log.value
    }
  }

  dynamic "enabled_log" {
    for_each = var.category_groups
    content {
      category_group = enabled_log.value
    }
  }

  dynamic "enabled_metric" {
    for_each = var.enable_metrics ? ["AllMetrics"] : []
    content {
      category = "AllMetrics"
    }
  }
}
