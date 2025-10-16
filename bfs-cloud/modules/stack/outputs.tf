output "resource_group_name" {
  description = "Name of the created resource group"
  value       = module.rg.name
}

output "vnet_id" {
  description = "ID of the virtual network"
  value       = module.net.vnet_id
}

output "subnet_id" {
  description = "ID of the container apps subnet"
  value       = module.net.subnet_id
}

output "container_app_environment_id" {
  description = "ID of the container app environment"
  value       = module.aca_env.id
}

output "log_analytics_workspace_id" {
  description = "ID of the Log Analytics workspace"
  value       = module.obs.log_analytics_id
}

output "app_insights_connection_string" {
  description = "Application Insights connection string"
  value       = var.config.enable_app_insights ? module.obs.app_insights_connection_string : null
  sensitive   = true
}

output "cosmos_connection_string" {
  description = "Cosmos DB connection string"
  value       = module.cosmos.connection_string
  sensitive   = true
}

output "app_urls" {
  description = "URLs of the deployed applications"
  value       = { for k, m in module.apps : k => m.url }
}

output "key_vault_secret_ids" {
  description = "Map of Key Vault secret names to their versionless IDs"
  value       = var.config.enable_security_features ? module.security[0].key_vault_secret_ids : {}
}

