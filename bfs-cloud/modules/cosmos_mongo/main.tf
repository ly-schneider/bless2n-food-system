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

  public_network_access_enabled     = true
  is_virtual_network_filter_enabled = false

  ip_range_filter = var.allowed_ip_ranges

  local_authentication_disabled = var.disable_local_auth

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

resource "azurerm_private_dns_zone" "cosmos_mongo" {
  name                = "privatelink.mongo.cosmos.azure.com"
  resource_group_name = var.resource_group_name
  tags                = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "cosmos_mongo_link" {
  name                  = "${var.name}-mongo-vnet-link"
  resource_group_name   = var.resource_group_name
  private_dns_zone_name = azurerm_private_dns_zone.cosmos_mongo.name
  virtual_network_id    = var.vnet_id
  registration_enabled  = false
}

resource "azurerm_private_endpoint" "cosmos_mongo" {
  name                = "${var.name}-pe"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.private_endpoint_subnet_id
  tags                = var.tags

  private_service_connection {
    name                           = "${var.name}-mongo-psc"
    private_connection_resource_id = azurerm_cosmosdb_account.this.id
    is_manual_connection           = false
    subresource_names              = ["MongoDB"]
  }

  private_dns_zone_group {
    name                 = "cosmos-mongo-dns-group"
    private_dns_zone_ids = [azurerm_private_dns_zone.cosmos_mongo.id]
  }
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
