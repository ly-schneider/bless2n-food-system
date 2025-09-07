resource "azurerm_container_app" "this" {
  name                         = var.name
  resource_group_name          = var.resource_group_name
  container_app_environment_id = var.environment_id
  tags                         = var.tags

  revision_mode = "Single"

  ingress {
    external_enabled = true
    target_port      = var.target_port
    transport        = "auto"

    traffic_weight {
      latest_revision = true
      percentage      = 100
    }
  }

  template {
    min_replicas = var.min_replicas
    max_replicas = var.max_replicas

    container {
      name   = "app"
      image  = var.image
      cpu    = var.cpu
      memory = var.memory

      dynamic "env" {
        for_each = var.environment_variables
        content {
          name  = env.key
          value = env.value
        }
      }
    }
  }

  identity {
    type = var.enable_system_identity ? "SystemAssigned" : "None"
  }
}

module "diag" {
  count                       = var.log_analytics_workspace_id == null ? 0 : 1
  source                      = "../diagnostic_setting"
  target_resource_id          = azurerm_container_app.this.id
  name                        = "${var.name}-diag"
  log_analytics_workspace_id  = var.log_analytics_workspace_id
  categories                  = []
  category_groups             = ["allLogs"]
  enable_metrics              = true
}