output "log_analytics_id"   { value = azurerm_log_analytics_workspace.law.id }
output "log_analytics_name" { value = azurerm_log_analytics_workspace.law.name }

output "app_insights_id" {
  value       = try(azurerm_application_insights.appi[0].id, null)
  description = "Null when disabled."
}

output "app_insights_connection_string" {
  value       = try(azurerm_application_insights.appi[0].connection_string, null)
  description = "Use with SDKs if needed."
}