resource "azurerm_consumption_budget_resource_group" "this" {
  name              = var.name
  resource_group_id = var.resource_group_id
  amount            = var.amount
  time_grain        = var.time_grain

  time_period {
    start_date = var.start_date
  }

  // 75% actual spend notification to Action Group (preferred) and Owners (fallback)
  notification {
    enabled        = true
    threshold      = 75
    operator       = "GreaterThan"
    threshold_type = "Actual"
    contact_groups = var.action_group_id == null ? [] : [var.action_group_id]
    contact_roles  = ["Owner"]
  }

  // 100% actual spend notification (hard stop alert)
  notification {
    enabled        = true
    threshold      = 100
    operator       = "GreaterThan"
    threshold_type = "Actual"
    contact_groups = var.action_group_id == null ? [] : [var.action_group_id]
    contact_roles  = ["Owner"]
  }
}
