output "id" { value = azurerm_container_app.this.id }
output "name" { value = azurerm_container_app.this.name }
output "url" { value = azurerm_container_app.this.latest_revision_fqdn }
output "fqdn" { value = azurerm_container_app.this.ingress[0].fqdn }
output "identity_principal_id" { value = var.enable_system_identity ? azurerm_container_app.this.identity[0].principal_id : null }
output "custom_domain_verification_id" { value = azurerm_container_app.this.custom_domain_verification_id }
