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
    min_replicas    = var.min_replicas
    max_replicas    = var.max_replicas
    revision_suffix = var.revision_suffix

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

    dynamic "http_scale_rule" {
      for_each = var.http_scale_rule != null ? [var.http_scale_rule] : []
      content {
        name                = http_scale_rule.value.name
        concurrent_requests = http_scale_rule.value.concurrent_requests
      }
    }

    dynamic "azure_queue_scale_rule" {
      for_each = var.azure_queue_scale_rules
      content {
        name         = azure_queue_scale_rule.value.name
        queue_name   = azure_queue_scale_rule.value.queue_name
        queue_length = azure_queue_scale_rule.value.queue_length
        
        authentication {
          secret_name       = azure_queue_scale_rule.value.secret_name
          trigger_parameter = azure_queue_scale_rule.value.trigger_parameter
        }
      }
    }

    dynamic "cpu_scale_rule" {
      for_each = var.cpu_scale_rule != null ? [var.cpu_scale_rule] : []
      content {
        name                = cpu_scale_rule.value.name
        cpu_percentage      = cpu_scale_rule.value.cpu_percentage
      }
    }

    dynamic "memory_scale_rule" {
      for_each = var.memory_scale_rule != null ? [var.memory_scale_rule] : []
      content {
        name               = memory_scale_rule.value.name
        memory_percentage  = memory_scale_rule.value.memory_percentage
      }
    }

    dynamic "custom_scale_rule" {
      for_each = var.custom_scale_rules
      content {
        name             = custom_scale_rule.value.name
        custom_rule_type = custom_scale_rule.value.custom_rule_type
        metadata         = custom_scale_rule.value.metadata

        dynamic "authentication" {
          for_each = custom_scale_rule.value.authentication != null ? [custom_scale_rule.value.authentication] : []
          content {
            secret_name       = authentication.value.secret_name
            trigger_parameter = authentication.value.trigger_parameter
          }
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