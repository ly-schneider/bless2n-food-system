locals {
  frontend_image = "${var.registry_server}/${var.registry_namespace}/frontend:${var.image_tag}"
  backend_image  = "${var.registry_server}/${var.registry_namespace}/backend:${var.image_tag}"
}

module "bfs_infrastructure" {
  source = "../../modules/stack"

  environment   = "production"
  location      = var.location
  tags          = var.tags
  alert_emails  = var.alert_emails

  config = {
    rg_name                      = "bfs-production-rg"
    vnet_name                    = "bfs-production-vnet"
    subnet_name                  = "container-apps-subnet"
    vnet_cidr                    = "10.2.0.0/16"
    subnet_cidr                  = "10.2.0.0/21"
    env_name                     = "bfs-production-env"
    law_name                     = "bfs-logs-workspace"
    appi_name                    = "bfs-production-insights"
    enable_app_insights          = true
    retention_days               = 30
    cosmos_name                  = "bfs-production-cosmos"
    database_throughput          = 800
    enable_alerts                = true
    requests_5xx_threshold       = 30
    enable_security_features     = true
    key_vault_name               = "bfs-production-kv"
    
    apps = {
      frontend-production = {
        port         = 80
        image        = local.frontend_image
        cpu          = 0.5
        memory       = "1.0Gi"
        min_replicas = 1
        max_replicas = 20
        registries = concat(
          var.registry_token != null && var.registry_username != null ? [{
            server                = var.registry_server
            username              = var.registry_username
            password_secret_name  = "ghcr-token"
          }] : [],
          lookup(var.app_registries, "frontend-production", [])
        )
        secrets = merge(
          var.registry_token != null ? { "ghcr-token" = var.registry_token } : {},
          lookup(var.app_secrets, "frontend-production", {})
        )
        http_scale_rule = {
          name                = "frontend-http-scale"
          concurrent_requests = 30
        }
        cpu_scale_rule = {
          name           = "frontend-cpu-scale"
          cpu_percentage = 70
        }
      }
      backend-production = {
        port         = 8080
        image        = local.backend_image
        cpu          = 0.75
        memory       = "1.5Gi"
        min_replicas = 1
        max_replicas = 20
        registries = concat(
          var.registry_token != null && var.registry_username != null ? [{
            server                = var.registry_server
            username              = var.registry_username
            password_secret_name  = "ghcr-token"
          }] : [],
          lookup(var.app_registries, "backend-production", [])
        )
        secrets = merge(
          var.registry_token != null ? { "ghcr-token" = var.registry_token } : {},
          lookup(var.app_secrets, "backend-production", {})
        )
        http_scale_rule = {
          name                = "backend-http-scale"
          concurrent_requests = 50
        }
        cpu_scale_rule = {
          name           = "backend-cpu-scale"
          cpu_percentage = 80
        }
        memory_scale_rule = {
          name              = "backend-mem-scale"
          memory_percentage = 85
        }
      }
    }
  }
}
