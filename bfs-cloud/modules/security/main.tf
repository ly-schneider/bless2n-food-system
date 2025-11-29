resource "azurerm_network_security_group" "aca_nsg" {
  name                = "${var.name_prefix}-aca-nsg"
  location            = var.location
  resource_group_name = var.resource_group_name

  # Allow HTTPS
  security_rule {
    name                       = "AllowHTTPS"
    priority                   = 1001
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  # Allow HTTP (for health checks and basic access)
  security_rule {
    name                       = "AllowHTTP"
    priority                   = 1002
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "80"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  # Allow Container Apps management traffic
  security_rule {
    name                       = "AllowContainerAppsInbound"
    priority                   = 1003
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_ranges    = ["5671", "5672"]
    source_address_prefix      = "AzureCloud"
    destination_address_prefix = "*"
  }

  # Allow all outbound traffic for Container Apps
  security_rule {
    name                       = "AllowContainerAppsOutbound"
    priority                   = 1004
    direction                  = "Outbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }


  tags = var.tags
}

resource "azurerm_subnet_network_security_group_association" "aca_nsg_association" {
  subnet_id                 = var.subnet_id
  network_security_group_id = azurerm_network_security_group.aca_nsg.id
}

resource "azurerm_key_vault" "basic" {
  count = var.enable_key_vault ? 1 : 0

  name                       = var.key_vault_name
  location                   = var.location
  resource_group_name        = var.resource_group_name
  tenant_id                  = data.azurerm_client_config.current.tenant_id
  sku_name                   = "standard"
  soft_delete_retention_days = 7

  public_network_access_enabled = true
  # CHANGED: Use Access Policies instead of RBAC to avoid needing User Access Administrator role
  enable_rbac_authorization = false

  network_acls {
    default_action = "Allow"
    bypass         = "AzureServices"
  }

  tags = var.tags
}

data "azurerm_client_config" "current" {}

# REMOVED: Role assignments for Key Vault (would require User Access Administrator role)
# Using Access Policies instead (defined in azurerm_key_vault resource above)

# Access Policy for Terraform service principal (admin access)
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

# Access Policy for Container Apps Managed Identity (read-only access to secrets)
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
  count = var.enable_key_vault && length(var.cosmos_connection_string) > 0 ? 1 : 0

  name         = "mongo-uri"
  value        = var.cosmos_connection_string
  key_vault_id = azurerm_key_vault.basic[0].id

  depends_on = [azurerm_key_vault_access_policy.terraform_admin]
}

resource "azurerm_monitor_metric_alert" "high_error_rate" {
  count = var.enable_basic_monitoring ? length(var.container_app_ids) : 0

  name                = "${var.name_prefix}-high-errors-${count.index}"
  resource_group_name = var.resource_group_name
  scopes              = [values(var.container_app_ids)[count.index]]
  description         = "High error rate detected"

  frequency   = "PT5M"
  window_size = "PT15M"
  severity    = 2

  criteria {
    metric_namespace = "Microsoft.App/containerapps"
    metric_name      = "Requests"
    aggregation      = "Count"
    operator         = "GreaterThan"
    threshold        = 50

    dimension {
      name     = "StatusCodeCategory"
      operator = "Include"
      values   = ["4xx", "5xx"]
    }
  }

  action {
    action_group_id = var.action_group_id
  }

  tags = var.tags
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
