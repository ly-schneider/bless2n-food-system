output "nsg_id" {
  description = "Network Security Group ID"
  value       = azurerm_network_security_group.aca_nsg.id
}

output "key_vault_id" {
  description = "Key Vault ID (if enabled)"
  value       = var.enable_key_vault ? azurerm_key_vault.basic[0].id : null
}

output "key_vault_uri" {
  description = "Key Vault URI (if enabled)"
  value       = var.enable_key_vault ? azurerm_key_vault.basic[0].vault_uri : null
}

output "key_vault_secret_ids" {
  description = "Map of Key Vault secret names to their versionless IDs"
  value = var.enable_key_vault && length(azurerm_key_vault_secret.cosmos_connection_string) > 0 ? {
    "mongo-uri" = azurerm_key_vault_secret.cosmos_connection_string[0].versionless_id
  } : {}
}

output "cost_savings_summary" {
  description = "Estimated monthly cost savings"
  value = {
    removed_features = [
      "Azure Security Center Standard (~$150/month)",
      "Web Application Firewall (~$40/month)",
      "Private Endpoints (~$30/month)",
      "Recovery Services Vault (~$20/month)",
      "Key Vault Premium upgrade (~$10/month)"
    ]
    total_estimated_savings = "~$250/month"
    retained_security = [
      "Network Security Groups (free)",
      "Managed Identities (free)",
      "Basic Key Vault Standard (~$1/month)",
      "Container security contexts (free)",
      "Basic monitoring alerts (~$2/month)"
    ]
    total_estimated_cost = "~$3/month"
  }
}
