locals {
  registry_host = var.enable_acr && var.acr_name != null && var.acr_name != "" ? "${var.acr_name}.azurecr.io" : var.registry_server
  frontend_repo = var.enable_acr && var.acr_name != null && var.acr_name != "" ? "frontend" : "${var.registry_namespace}/frontend"
  backend_repo  = var.enable_acr && var.acr_name != null && var.acr_name != "" ? "backend"  : "${var.registry_namespace}/backend"

  frontend_image = "${local.registry_host}/${local.frontend_repo}:${var.image_tag}"
  backend_image  = "${local.registry_host}/${local.backend_repo}:${var.image_tag}"
}

module "bfs_infrastructure" {
  source = "../../modules/stack"

  environment   = "staging"
  location      = var.location
  tags          = var.tags
  alert_emails  = var.alert_emails

  config = {
    rg_name                      = "bfs-staging-rg"
    vnet_name                    = "bfs-staging-vnet"
    subnet_name                  = "container-apps-subnet"
    vnet_cidr                    = "10.1.0.0/16"
    subnet_cidr                  = "10.1.0.0/21"
    env_name                     = "bfs-staging-env"
    law_name                     = "bfs-logs-workspace"
    appi_name                    = "bfs-staging-insights"
    enable_app_insights          = false
    retention_days               = 30
    cosmos_name                  = "bfs-staging-cosmos"
    database_throughput          = 400
    enable_alerts                = false
    requests_5xx_threshold       = 10
    enable_security_features     = true
    key_vault_name               = "bfs-staging-kv"
    enable_acr                   = true
    acr_name                     = "bfsstagingacr"
    
    apps = {
      frontend-staging = {
        port         = 80
        image        = local.frontend_image
        cpu          = 0.25
        memory       = "0.5Gi"
        min_replicas = 0
        max_replicas = 20
        # Optional: registry and secrets for private images (GHCR) plus per-app overrides
        registries = var.enable_acr ? lookup(var.app_registries, "frontend-staging", []) : concat(
          var.registry_token != null && var.registry_username != null ? [{
            server                = var.registry_server
            username              = var.registry_username
            password_secret_name  = "ghcr-token"
          }] : [],
          lookup(var.app_registries, "frontend-staging", [])
        )
        secrets = var.enable_acr ? lookup(var.app_secrets, "frontend-staging", {}) : merge(
          var.registry_token != null ? { "ghcr-token" = var.registry_token } : {},
          lookup(var.app_secrets, "frontend-staging", {})
        )
        http_scale_rule = {
          name                = "frontend-http-scale"
          concurrent_requests = 20
        }
        cpu_scale_rule = {
          name           = "frontend-cpu-scale"
          cpu_percentage = 75
        }
      }
      backend-staging = {
        port         = 8080
        image        = local.backend_image
        cpu          = 0.5
        memory       = "1.0Gi"
        min_replicas = 0
        max_replicas = 20
        registries = var.enable_acr ? lookup(var.app_registries, "backend-staging", []) : concat(
          var.registry_token != null && var.registry_username != null ? [{
            server                = var.registry_server
            username              = var.registry_username
            password_secret_name  = "ghcr-token"
          }] : [],
          lookup(var.app_registries, "backend-staging", [])
        )
        secrets = var.enable_acr ? lookup(var.app_secrets, "backend-staging", {}) : merge(
          var.registry_token != null ? { "ghcr-token" = var.registry_token } : {},
          lookup(var.app_secrets, "backend-staging", {})
        )
        http_scale_rule = {
          name                = "backend-http-scale"
          concurrent_requests = 40
        }
        cpu_scale_rule = {
          name           = "backend-cpu-scale"
          cpu_percentage = 80
        }
      }
    }
  }
}
