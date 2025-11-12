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
  description = "URLs of the deployed applications (latest revision FQDNs)"
  value = jsonencode(merge(
    { for k, m in module.apps_backend : k => m.url },
    { for k, m in module.apps_frontend : k => m.url }
  ))
}

output "backend_fqdns" {
  description = "Stable FQDNs (ingress.fqdn) for backend apps"
  value       = jsonencode({ for k, m in module.apps_backend : k => m.fqdn })
}

output "key_vault_id" {
  description = "Key Vault ID"
  value       = var.config.enable_security_features ? module.security[0].key_vault_id : null
}

output "key_vault_secret_ids" {
  description = "Map of Key Vault secret names to their versionless IDs"
  value       = jsonencode(var.config.enable_security_features ? module.security[0].key_vault_secret_ids : {})
}

output "dns_zone_name" {
  description = "Public DNS zone name if created (e.g., food.bless2n.ch)"
  value       = length(azurerm_dns_zone.public) > 0 ? azurerm_dns_zone.public[0].name : null
}

output "dns_zone_name_servers" {
  description = "Nameservers for the public DNS zone (delegate at parent provider)"
  value       = length(azurerm_dns_zone.public) > 0 ? azurerm_dns_zone.public[0].name_servers : null
}

output "domain_verification_ids" {
  description = "Map of app name -> custom_domain_verification_id"
  value = jsonencode(merge(
    { for k, m in module.apps_backend : k => m.custom_domain_verification_id },
    { for k, m in module.apps_frontend : k => m.custom_domain_verification_id }
  ))
}
