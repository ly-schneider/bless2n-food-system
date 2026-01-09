# Cosmos DB with MongoDB API - public network access (no VNet integration)
resource "azurerm_cosmosdb_account" "this" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name

  offer_type = "Standard"
  kind       = "MongoDB"

  mongo_server_version = "7.0"

  capabilities {
    name = "EnableMongo"
  }

  capabilities {
    name = "EnableServerless"
  }

  consistency_policy {
    consistency_level = "Session"
  }

  geo_location {
    location          = var.location
    failover_priority = 0
  }

  backup {
    type                = "Periodic"
    interval_in_minutes = 240
    retention_in_hours  = 8
    storage_redundancy  = "Geo"
  }

  public_network_access_enabled     = true
  is_virtual_network_filter_enabled = false

  # Only apply IP filtering if caller provided ranges; otherwise use empty set
  ip_range_filter = length(var.allowed_ip_ranges) > 0 ? toset(var.allowed_ip_ranges) : []

  cors_rule {
    allowed_origins    = var.cors_allowed_origins
    allowed_methods    = ["GET", "POST"]
    allowed_headers    = ["*"]
    exposed_headers    = ["*"]
    max_age_in_seconds = 3600
  }

  identity {
    type = "SystemAssigned"
  }

  tags = var.tags
}

resource "azurerm_cosmosdb_mongo_database" "db" {
  count               = var.create_database ? 1 : 0
  name                = var.database_name
  resource_group_name = var.resource_group_name
  account_name        = azurerm_cosmosdb_account.this.name
}

resource "azurerm_monitor_diagnostic_setting" "cosmos_diag" {
  name                       = "${var.name}-diag"
  target_resource_id         = azurerm_cosmosdb_account.this.id
  log_analytics_workspace_id = var.log_analytics_workspace_id

  enabled_log {
    category = "DataPlaneRequests"
  }

  enabled_log {
    category = "MongoRequests"
  }

  enabled_log {
    category = "QueryRuntimeStatistics"
  }

  enabled_log {
    category = "PartitionKeyStatistics"
  }

  enabled_metric {
    category = "Requests"
  }
}

# Ensure the deploying principal can read/manage Cosmos keys so provider
# operations (like listKeys) succeed during create/read.
data "azurerm_client_config" "current" {}
