resource "azurerm_storage_account" "this" {
  name                     = var.name
  resource_group_name      = var.resource_group_name
  location                 = var.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  account_kind             = "StorageV2"
  min_tls_version          = "TLS1_2"

  blob_properties {
    cors_rule {
      allowed_headers    = ["*"]
      allowed_methods    = ["GET"]
      allowed_origins    = ["*"]
      exposed_headers    = ["Content-Length", "Content-Type"]
      max_age_in_seconds = 3600
    }
  }

  tags = var.tags
}

resource "azurerm_storage_container" "images" {
  name                  = var.container_name
  storage_account_id    = azurerm_storage_account.this.id
  container_access_type = "blob"
}
