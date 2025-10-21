// Action Group is provided by the caller (stack) to avoid count-on-unknown problems

resource "azurerm_monitor_metric_alert" "requests_5xx" {
  for_each            = var.container_app_ids
  name                = "${each.key}-5xx-rate"
  resource_group_name = var.resource_group_name
  scopes              = [each.value]
  description         = "High HTTP 5xx on ${each.key}"

  frequency     = "PT1M"
  window_size   = "PT5M"
  severity      = 1
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
    action_group_id = var.action_group_id
  }

  tags = var.tags
}
