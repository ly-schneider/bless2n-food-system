locals {
  registry_host = "ghcr.io"
  repo_prefix   = "ly-schneider/bless2n-food-system"
  frontend_repo = "frontend"
  backend_repo  = "backend"

  frontend_image = var.frontend_digest != "" ? "${local.registry_host}/${local.repo_prefix}/${local.frontend_repo}@${var.frontend_digest}" : "${local.registry_host}/${local.repo_prefix}/${local.frontend_repo}:${var.image_tag}"

  backend_image = var.backend_digest != "" ? "${local.registry_host}/${local.repo_prefix}/${local.backend_repo}@${var.backend_digest}" : "${local.registry_host}/${local.repo_prefix}/${local.backend_repo}:${var.image_tag}"
}

module "bfs_infrastructure" {
  source = "../../modules/stack"

  environment = "production"
  location    = var.location
  tags        = var.tags

  config = {
    rg_name                  = "bfs-production-rg"
    vnet_name                = "bfs-production-vnet"
    subnet_name              = "container-apps-subnet"
    vnet_cidr                = "10.1.0.0/16"
    subnet_cidr              = "10.1.0.0/21"
    delegate_aca_subnet      = false
    pe_subnet_name           = "private-endpoints-subnet"
    pe_subnet_cidr           = "10.1.8.0/24"
    env_name                 = "bfs-production-env"
    law_name                 = "bfs-logs-workspace"
    appi_name                = "bfs-production-insights"
    enable_app_insights      = false
    retention_days           = 30
    cosmos_name              = "bfs-production-cosmos"
    database_throughput      = 400
    enable_security_features = true
    enable_private_endpoint  = false
    key_vault_name           = "bfs-production-keyvault"
    apps = {
      frontend-production = {
        port                           = 3000
        image                          = local.frontend_image
        revision_suffix                = var.revision_suffix
        external_ingress               = true
        allow_insecure_connections     = true
        cpu                            = 0.5
        memory                         = "1Gi"
        min_replicas                   = 1
        max_replicas                   = 10
        health_check_path              = "/health"
        liveness_path                  = "/health"
        liveness_interval_seconds      = 30
        liveness_initial_delay_seconds = 20
        registries = [{
          server               = "ghcr.io"
          username             = "ly-schneider"
          password_secret_name = "ghcr-token"
        }]
        secrets = merge(
          lookup(var.app_secrets, "frontend-production", {}),
          {
            ghcr-token = var.ghcr_token
          }
        )
        environment_variables = {
          NODE_ENV = "production"

          LOG_LEVEL = "info"

          NEXT_PUBLIC_POS_PIN          = "0000"
          NEXT_PUBLIC_POS_IDLE_TIMEOUT = "300000"

          NEXT_PUBLIC_GA_MEASUREMENT_ID = "G-9W8S03MJEM"

          NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY = "pk_live_51PBsTW2LfBQkI29zWnUaY1HAsP34VeFoYsInuhNNpyHjrZiHhrDZrMGvCIzZuZOXJEzxpOXhQyW8wfi6qiwelHkh00TTmNdU4k"

          NEXT_PUBLIC_GOOGLE_CLIENT_ID = "728225904671-bsdo738sald74qkbg38etqjqj5jjfh66.apps.googleusercontent.com"
        }
        key_vault_secrets = merge(
          lookup(var.app_secrets, "frontend-production", {}),
          {
            "BACKEND_INTERNAL_URL" = "backend-internal-url"
          }
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
      backend-production = {
        port                           = 8080
        image                          = local.backend_image
        revision_suffix                = var.revision_suffix
        external_ingress               = true
        allow_insecure_connections     = true
        health_check_path              = "/health"
        liveness_path                  = "/ping"
        liveness_interval_seconds      = 60
        liveness_initial_delay_seconds = 30
        cpu                            = 0.5
        memory                         = "1Gi"
        min_replicas                   = 1
        max_replicas                   = 10
        registries = [{
          server               = "ghcr.io"
          username             = "ly-schneider"
          password_secret_name = "ghcr-token"
        }]
        secrets = merge(
          lookup(var.app_secrets, "backend-production", {}),
          {
            ghcr-token = var.ghcr_token
          }
        )
        environment_variables = {
          APP_ENV  = "production"
          APP_PORT = "8080"

          LOG_LEVEL       = "info"
          LOG_DEVELOPMENT = "false"

          SECURITY_ENABLE_HSTS = "true"
          SECURITY_ENABLE_CSP  = "true"

          SMTP_HOST       = "mail.infomaniak.com"
          SMTP_PORT       = "587"
          SMTP_USERNAME   = "levyn.schneider@rentro.ch"
          SMTP_FROM       = "BlessThun Food <levyn.schneider@rentro.ch>"
          SMTP_TLS_POLICY = "starttls"

          GOOGLE_CLIENT_ID = "728225904671-bsdo738sald74qkbg38etqjqj5jjfh66.apps.googleusercontent.com"

          MONGO_DATABASE = "bless2n_food_system"

          STATION_QR_MAX_AGE_SECONDS = "86400"
        }
        key_vault_secrets = merge(
          lookup(var.app_secrets, "backend-production", {}),
          {
            "MONGO_URI"                = "mongo-uri"
            "JWT_PRIV_PEM"             = "jwt-priv-pem"
            "JWT_PUB_PEM"              = "jwt-pub-pem"
            "STATION_QR_SECRET"        = "station-qr-secret"
            "GOOGLE_CLIENT_SECRET"     = "google-client-secret"
            "STRIPE_SECRET_KEY"        = "stripe-secret-key"
            "STRIPE_WEBHOOK_SECRET"    = "stripe-webhook-secret"
            "SMTP_PASSWORD"            = "smtp-password"
            "SECURITY_TRUSTED_ORIGINS" = "security-trusted-origins"
            "PUBLIC_BASE_URL"          = "public-base-url"
            "JWT_ISSUER"               = "jwt-issuer"
          }
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