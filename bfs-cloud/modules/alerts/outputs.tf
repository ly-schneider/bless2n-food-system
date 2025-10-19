output "action_group_id" {
  value = try(azurerm_monitor_action_group.this[0].id, var.action_group_id)
}
