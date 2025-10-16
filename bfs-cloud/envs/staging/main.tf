locals {
  registry_host = "${var.acr_name}.azurecr.io"
  frontend_repo = "frontend"
  backend_repo  = "backend"

  frontend_image = "${local.registry_host}/${local.frontend_repo}:${var.image_tag}"
  backend_image  = "${local.registry_host}/${local.backend_repo}:${var.image_tag}"
}

module "bfs_infrastructure" {
  source = "../../modules/stack"

  environment  = "staging"
  location     = var.location
  tags         = var.tags
  alert_emails = var.alert_emails

  config = {
    rg_name                  = "bfs-staging-rg"
    vnet_name                = "bfs-staging-vnet"
    subnet_name              = "container-apps-subnet"
    vnet_cidr                = "10.1.0.0/16"
    subnet_cidr              = "10.1.0.0/21"
    env_name                 = "bfs-staging-env"
    law_name                 = "bfs-logs-workspace"
    appi_name                = "bfs-staging-insights"
    enable_app_insights      = false
    retention_days           = 30
    cosmos_name              = "bfs-staging-cosmos"
    database_throughput      = 400
    enable_alerts            = false
    requests_5xx_threshold   = 10
    enable_security_features = true
    key_vault_name           = "bfs-staging-kv"
    acr_name                 = "bfsstagingacr"

    apps = {
      frontend-staging = {
        port             = 80
        image            = local.frontend_image
        external_ingress = true
        cpu              = 0.25
        memory           = "0.5Gi"
        min_replicas     = 0
        max_replicas     = 20
        # Use an internal health endpoint that does not depend on backend
        health_check_path = "/api/health"
        # ACR configuration - no additional registries needed when using managed identity
        registries = []
        secrets    = lookup(var.app_secrets, "frontend-staging", {})
        environment_variables = {
          # Application configuration
          NODE_ENV                 = "production"
          NEXT_PUBLIC_API_BASE_URL = "https://staging.food.blessthun.ch"
          BACKEND_INTERNAL_URL     = "http://backend-staging"
          LOG_LEVEL                = "info"

          # POS configuration
          NEXT_PUBLIC_POS_PIN          = "0000"
          NEXT_PUBLIC_POS_IDLE_TIMEOUT = "300000"

          # Analytics (to be configured)
          NEXT_PUBLIC_GA_MEASUREMENT_ID = "G-9W8S03MJEM"

          # Public keys (safe to be in environment variables)
          NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY = "pk_test_51RQl9ZRqpUl5qfpdqiu8nhU6h5N1YyEXWoOxqUsPb8UouZqPMubOZtESdFa4KHTWM71GhAbbddlS3a6aTFu1vIDe00p1DqQTG9"
          NEXT_PUBLIC_GOOGLE_CLIENT_ID       = "728225904671-h9cp0badsuvamscrn6k2lnkksiinld99.apps.googleusercontent.com"
        }
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
        port             = 8080
        image            = local.backend_image
        external_ingress = false
        cpu              = 0.5
        memory           = "1.0Gi"
        min_replicas     = 0
        max_replicas     = 20
        # ACR configuration - no additional registries needed when using managed identity
        registries = []
        secrets    = lookup(var.app_secrets, "backend-staging", {})
        environment_variables = {
          # Application configuration
          APP_ENV         = "staging"
          APP_PORT        = "8080"
          LOG_LEVEL       = "info"
          LOG_DEVELOPMENT = "false"

          # Security settings for staging
          SECURITY_ENABLE_HSTS     = "true"
          SECURITY_ENABLE_CSP      = "true"
          SECURITY_TRUSTED_ORIGINS = "https://staging.food.blessthun.ch"

          # SMTP settings (can be configured for staging)
          SMTP_HOST       = "your-smtp-host"
          SMTP_PORT       = "587"
          SMTP_USERNAME   = "staging@yourdomain.com"
          SMTP_FROM       = "noreply@yourdomain.com"
          SMTP_TLS_POLICY = "require"

          # JWT settings
          JWT_ISSUER      = "https://staging.food.blessthun.ch"
          PUBLIC_BASE_URL = "https://staging.food.blessthun.ch"

          # Google OAuth (to be configured)
          GOOGLE_CLIENT_ID = "728225904671-h9cp0badsuvamscrn6k2lnkksiinld99.apps.googleusercontent.com"

          # Database configuration (using Key Vault secrets)
          MONGO_DATABASE = "bless2n_food_system"
        }
        key_vault_secrets = {
          # These env vars will get their values from Key Vault secrets
          MONGO_URI             = "mongo-connection-string"
          JWT_PRIV_PEM_PATH     = "jwt-private-key"
          JWT_PUB_PEM_PATH      = "jwt-public-key"
          GOOGLE_CLIENT_SECRET  = "google-client-secret"
          STRIPE_SECRET_KEY     = "stripe-secret-key"
          STRIPE_WEBHOOK_SECRET = "stripe-webhook-secret"
          SMTP_PASSWORD         = "smtp-password"
        }
        # No explicit key_vault_secret_refs here; IDs are resolved inside the stack module from names
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
