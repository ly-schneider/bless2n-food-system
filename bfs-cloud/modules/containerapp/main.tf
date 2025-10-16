locals {
  kv_secret_identity = var.enable_system_identity && length(var.user_assigned_identity_ids) == 0 ? "System" : (length(var.user_assigned_identity_ids) > 0 ? var.user_assigned_identity_ids[0] : null)
  _kv_identity_guard = length(var.key_vault_secret_refs) == 0 || local.kv_secret_identity != null ? true : tomap({})["force_error"]
}

resource "azurerm_container_app" "this" {
  name                         = var.name
  resource_group_name          = var.resource_group_name
  container_app_environment_id = var.environment_id
  tags                         = var.tags

  revision_mode = "Single"

  # Optional secrets for use by registries or app config
  dynamic "secret" {
    for_each = var.secrets
    content {
      name  = secret.key
      value = secret.value
    }
  }

  # Key Vault secret references
  dynamic "secret" {
    for_each = var.key_vault_secret_refs
    content {
      name                = secret.key
      key_vault_secret_id = secret.value
      # Required by AzureRM when referencing Key Vault secrets
      identity            = local.kv_secret_identity
    }
  }

  # Optional container registry credentials (e.g., GHCR)
  dynamic "registry" {
    for_each = var.registries
    content {
      server               = registry.value.server
      # Support username/password-based registries (e.g., GHCR)
      username             = try(registry.value.username, null)
      password_secret_name = try(registry.value.password_secret_name, null)
      # Support managed identity-based ACR pull
      identity             = try(registry.value.identity, null)
    }
  }

  ingress {
    external_enabled           = var.external_ingress
    target_port                = var.target_port
    transport                  = "auto"
    allow_insecure_connections = false

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

      dynamic "env" {
        for_each = var.key_vault_secrets
        content {
          name        = env.key
          secret_name = env.value
        }
      }


      liveness_probe {
        path      = var.health_check_path
        port      = var.target_port
        transport = "HTTP"
      }

      readiness_probe {
        path      = var.health_check_path
        port      = var.target_port
        transport = "HTTP"
      }

      dynamic "volume_mounts" {
        for_each = var.volume_mounts
        content {
          name = volume_mounts.value.name
          path = volume_mounts.value.path
        }
      }
    }

    dynamic "volume" {
      for_each = var.volumes
      content {
        name         = volume.value.name
        storage_type = volume.value.storage_type
        storage_name = volume.value.storage_name
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

    dynamic "custom_scale_rule" {
      for_each = var.cpu_scale_rule != null ? [var.cpu_scale_rule] : []
      content {
        name             = custom_scale_rule.value.name
        custom_rule_type = "cpu"
        metadata = {
          type  = "Utilization"
          value = tostring(custom_scale_rule.value.cpu_percentage)
        }
      }
    }

    dynamic "custom_scale_rule" {
      for_each = var.memory_scale_rule != null ? [var.memory_scale_rule] : []
      content {
        name             = memory_scale_rule.value.name
        custom_rule_type = "memory"
        metadata = {
          type  = "Utilization"
          value = tostring(memory_scale_rule.value.memory_percentage)
        }
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
    type         = var.enable_system_identity && length(var.user_assigned_identity_ids) == 0 ? "SystemAssigned" : length(var.user_assigned_identity_ids) > 0 ? "UserAssigned" : "None"
    identity_ids = length(var.user_assigned_identity_ids) > 0 ? var.user_assigned_identity_ids : null
  }
}

module "diag" {
  source                     = "../diagnostic_setting"
  target_resource_id         = azurerm_container_app.this.id
  name                       = "${var.name}-diag"
  log_analytics_workspace_id = var.log_analytics_workspace_id
  categories                 = []
  category_groups            = ["allLogs"]
  enable_metrics             = true
}
