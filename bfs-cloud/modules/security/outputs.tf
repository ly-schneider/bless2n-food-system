# Outputs reference the existing Key Vault via data source
# The Key Vault is no longer managed by Terraform but can still be referenced

output "key_vault_id" {
  description = "Key Vault ID (if enabled)"
  value       = var.enable_key_vault ? data.azurerm_key_vault.existing[0].id : null
}

output "key_vault_uri" {
  description = "Key Vault URI (if enabled)"
  value       = var.enable_key_vault ? data.azurerm_key_vault.existing[0].vault_uri : null
}

output "key_vault_secret_ids" {
  description = "Map of Key Vault secret names to their versionless IDs (constructed from URI)"
  value       = var.enable_key_vault ? {} : {}
}
