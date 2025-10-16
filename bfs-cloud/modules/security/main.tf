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
  
  name                     = var.key_vault_name
  location                 = var.location
  resource_group_name      = var.resource_group_name
  tenant_id                = data.azurerm_client_config.current.tenant_id
  sku_name                 = "standard"
  soft_delete_retention_days = 7
  
  public_network_access_enabled = true
  rbac_authorization_enabled     = true

  network_acls {
    default_action = "Allow"
    bypass         = "AzureServices"
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
  count = var.enable_key_vault ? 1 : 0
  
  scope                = azurerm_key_vault.basic[0].id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = var.uami_principal_id
}

resource "azurerm_key_vault_secret" "cosmos_connection_string" {
  count = var.enable_key_vault ? 1 : 0
  
  name         = "mongo-connection-string"
  value        = var.cosmos_connection_string
  key_vault_id = azurerm_key_vault.basic[0].id

  depends_on = [azurerm_role_assignment.kv_admin]
}

# Placeholder secrets for staging - these should be populated manually or via CI/CD
resource "azurerm_key_vault_secret" "placeholder_secrets" {
  for_each = var.enable_key_vault ? {
    "jwt-private-key"              = "placeholder-jwt-private-key"
    "jwt-public-key"               = "placeholder-jwt-public-key"
    "google-client-secret"         = "placeholder-google-client-secret"
    "stripe-secret-key"            = "placeholder-stripe-secret-key"
    "stripe-webhook-secret"        = "placeholder-stripe-webhook-secret"
    "smtp-password"                = "placeholder-smtp-password"
  } : {}
  
  name         = each.key
  value        = each.value
  key_vault_id = azurerm_key_vault.basic[0].id

  depends_on = [azurerm_role_assignment.kv_admin]
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
  
  name               = "${var.key_vault_name}-basic-diag"
  target_resource_id = azurerm_key_vault.basic[0].id
  log_analytics_workspace_id = var.log_analytics_workspace_id

  enabled_log {
    category = "AuditEvent"
  }
}