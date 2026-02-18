# Key Vault resources have been removed from Terraform state management.
# The Key Vault continues to exist in Azure but is no longer managed by Terraform.
# See removed.tf for the removal configuration.
#
# To reference the existing Key Vault, use a data source:
data "azurerm_key_vault" "existing" {
  count = var.enable_key_vault ? 1 : 0

  name                = var.key_vault_name
  resource_group_name = var.resource_group_name
}

data "azurerm_client_config" "current" {}
