resource "azurerm_monitor_action_group" "this" {
  name                = var.name
  resource_group_name = var.resource_group_name
  short_name          = var.short_name

  dynamic "email_receiver" {
    for_each = var.email_receivers
    content {
      name          = email_receiver.key
      email_address = email_receiver.value
    }
  }

  tags = var.tags
}

resource "azurerm_monitor_metric_alert" "requests_5xx" {
  for_each            = var.container_app_ids
  name                = "${each.key}-5xx-rate"
  resource_group_name = var.resource_group_name
  scopes              = [each.value]
  description         = "High HTTP 5xx on ${each.key}"

  frequency   = "PT1M"
  window_size = "PT5M"
  severity    = 2
  auto_mitigate = true

  criteria {
    metric_namespace = "Microsoft.App/containerapps"
    metric_name      = "Requests"
    aggregation      = "Total"
    operator         = "GreaterThan"
    threshold        = var.requests_5xx_threshold

    dimension {
      name     = "StatusCodeCategory"
      operator = "Include"
      values   = ["5xx"]
    }
  }

  action {
    action_group_id = azurerm_monitor_action_group.this.id
  }

  tags = var.tags
}