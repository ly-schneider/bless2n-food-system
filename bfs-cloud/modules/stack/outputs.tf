output "resource_group_name" {
  description = "Name of the created resource group"
  value       = module.rg.name
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

output "app_urls" {
  description = "URLs of the deployed applications (latest revision FQDNs)"
  value = jsonencode(merge(
    { for k, m in module.apps_backend : k => m.url },
    { for k, m in module.apps_frontend : k => m.url },
    { for k, m in module.apps_docs : k => m.url }
  ))
}

output "backend_fqdns" {
  description = "Stable FQDNs (ingress.fqdn) for backend apps"
  value       = jsonencode({ for k, m in module.apps_backend : k => m.fqdn })
}

output "key_vault_id" {
  description = "Key Vault ID"
  value       = module.security.key_vault_id
}

output "key_vault_secret_ids" {
  description = "Map of Key Vault secret names to their versionless IDs"
  value       = jsonencode(module.security.key_vault_secret_ids)
}

output "blob_storage_endpoint" {
  description = "Primary blob storage endpoint"
  value       = module.blob_storage.blob_endpoint
}

output "frontend_fqdns" {
  description = "Stable FQDNs (ingress.fqdn) for frontend apps"
  value       = jsonencode({ for k, m in module.apps_frontend : k => m.fqdn })
}

output "docs_fqdns" {
  description = "Stable FQDNs (ingress.fqdn) for docs apps"
  value       = jsonencode({ for k, m in module.apps_docs : k => m.fqdn })
}
