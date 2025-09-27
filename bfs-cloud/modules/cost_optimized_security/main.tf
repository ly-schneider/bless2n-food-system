# Cost-Optimized Security Module
# Provides essential security at minimal cost

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

  # Block all other inbound traffic
  security_rule {
    name                       = "DenyAllOtherInbound"
    priority                   = 4000
    direction                  = "Inbound"
    access                     = "Deny"
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

# Basic Key Vault (Standard SKU - much cheaper than Premium)
resource "azurerm_key_vault" "basic" {
  count = var.enable_key_vault ? 1 : 0
  
  name                     = var.key_vault_name
  location                 = var.location
  resource_group_name      = var.resource_group_name
  tenant_id                = data.azurerm_client_config.current.tenant_id
  sku_name                 = "standard"  # Much cheaper than premium
  soft_delete_retention_days = 7
  
  # Public access enabled to avoid private endpoint costs
  public_network_access_enabled = true
  rbac_authorization_enabled     = true

  # Allow access from Azure services and specific IPs only
  network_acls {
    default_action = "Deny"
    bypass         = "AzureServices"
    ip_rules       = var.allowed_ip_ranges
  }

  tags = var.tags
}

data "azurerm_client_config" "current" {}

resource "azurerm_role_assignment" "kv_admin" {
  count = var.enable_key_vault ? length(var.key_vault_admins) : 0
  
  scope                = azurerm_key_vault.basic[0].id
  role_definition_name = "Key Vault Administrator"
  principal_id         = var.key_vault_admins[count.index]
}

resource "azurerm_role_assignment" "kv_secrets_user" {
  for_each = var.enable_key_vault ? var.container_app_identities : {}
  
  scope                = azurerm_key_vault.basic[0].id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = each.value
}

# Store only critical secrets
resource "azurerm_key_vault_secret" "cosmos_connection_string" {
  for_each = var.enable_key_vault ? { "cosmos" = var.cosmos_connection_string } : {}
  
  name         = "cosmos-connection-string"
  value        = each.value
  key_vault_id = azurerm_key_vault.basic[0].id

  depends_on = [azurerm_role_assignment.kv_admin]
}

# Basic monitoring alerts (using existing action group)
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

# Basic diagnostic logging to existing Log Analytics
resource "azurerm_monitor_diagnostic_setting" "basic_security_logs" {
  count = var.enable_key_vault ? 1 : 0
  
  name               = "${var.key_vault_name}-basic-diag"
  target_resource_id = azurerm_key_vault.basic[0].id
  log_analytics_workspace_id = var.log_analytics_workspace_id

  enabled_log {
    category = "AuditEvent"
  }
}