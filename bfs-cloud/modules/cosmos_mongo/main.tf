resource "azurerm_cosmosdb_account" "this" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name

  offer_type = "Standard"
  kind       = "MongoDB"

  capabilities {
    name = "EnableMongo"
  }

  capabilities {
    name = "MongoDBv3.4"
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

  public_network_access_enabled = true
  is_virtual_network_filter_enabled = true

  virtual_network_rule {
    id                                   = var.subnet_id
    ignore_missing_vnet_service_endpoint = false
  }

  ip_range_filter = var.allowed_ip_ranges

  local_authentication_disabled = false

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
  throughput          = var.database_throughput
}

# Private endpoints removed for cost optimization
# Access is secured through VNet service endpoints and IP restrictions

resource "azurerm_monitor_diagnostic_setting" "cosmos_diag" {
  name               = "${var.name}-diag"
  target_resource_id = azurerm_cosmosdb_account.this.id
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