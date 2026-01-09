resource "azurerm_key_vault" "basic" {
  count = var.enable_key_vault ? 1 : 0

  name                       = var.key_vault_name
  location                   = var.location
  resource_group_name        = var.resource_group_name
  tenant_id                  = data.azurerm_client_config.current.tenant_id
  sku_name                   = "standard"
  soft_delete_retention_days = 7

  public_network_access_enabled = true
  rbac_authorization_enabled    = false

  network_acls {
    default_action = "Allow"
    bypass         = "AzureServices"
  }

  tags = var.tags
}

data "azurerm_client_config" "current" {}

resource "azurerm_key_vault_access_policy" "terraform_admin" {
  count = var.enable_key_vault ? length(var.key_vault_admins) : 0

  key_vault_id = azurerm_key_vault.basic[0].id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = var.key_vault_admins[count.index]

  secret_permissions = [
    "Get",
    "List",
    "Set",
    "Delete",
    "Purge",
    "Recover",
    "Backup",
    "Restore"
  ]

  key_permissions = [
    "Get",
    "List",
    "Create",
    "Delete",
    "Purge",
    "Recover"
  ]

  certificate_permissions = [
    "Get",
    "List",
    "Create",
    "Delete",
    "Purge",
    "Recover"
  ]
}

resource "azurerm_key_vault_access_policy" "container_apps_identity" {
  count = var.enable_key_vault ? 1 : 0

  key_vault_id = azurerm_key_vault.basic[0].id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = var.uami_principal_id

  secret_permissions = [
    "Get",
    "List"
  ]
}

resource "azurerm_key_vault_secret" "cosmos_connection_string" {
  count = var.enable_key_vault ? 1 : 0

  name         = "mongo-uri"
  value        = var.cosmos_connection_string
  key_vault_id = azurerm_key_vault.basic[0].id

  depends_on = [azurerm_key_vault_access_policy.terraform_admin]
}

resource "azurerm_monitor_diagnostic_setting" "basic_security_logs" {
  count = var.enable_key_vault ? 1 : 0

  name                       = "${var.key_vault_name}-basic-diag"
  target_resource_id         = azurerm_key_vault.basic[0].id
  log_analytics_workspace_id = var.log_analytics_workspace_id

  enabled_log {
    category = "AuditEvent"
  }
}
