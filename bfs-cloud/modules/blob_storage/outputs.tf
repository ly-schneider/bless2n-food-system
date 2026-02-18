output "storage_account_name" {
  description = "Name of the storage account"
  value       = azurerm_storage_account.this.name
}

output "primary_access_key" {
  description = "Primary access key for the storage account"
  value       = azurerm_storage_account.this.primary_access_key
  sensitive   = true
}

output "blob_endpoint" {
  description = "Primary blob service endpoint"
  value       = azurerm_storage_account.this.primary_blob_endpoint
}

output "container_name" {
  description = "Name of the blob container"
  value       = azurerm_storage_container.images.name
}
