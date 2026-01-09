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
  value = var.enable_key_vault ? {
    "mongo-uri" = azurerm_key_vault_secret.cosmos_connection_string[0].versionless_id
  } : {}
}
