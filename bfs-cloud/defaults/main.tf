# Centralized configuration module
# Takes environment name and optional per-env inputs, outputs complete ready-to-use config

variable "env" {
  description = "Environment name (staging, production)"
  type        = string
}

variable "image_tag" {
  description = "Docker image tag"
  type        = string
}

variable "frontend_digest" {
  description = "Frontend image digest (optional, preferred over tag)"
  type        = string
  default     = ""
}

variable "backend_digest" {
  description = "Backend image digest (optional, preferred over tag)"
  type        = string
  default     = ""
}

variable "revision_suffix" {
  description = "Container Apps revision suffix"
  type        = string
  default     = null
}

variable "ghcr_token" {
  description = "GitHub Container Registry token"
  type        = string
  sensitive   = true
}

variable "app_secrets" {
  description = "Additional app secrets"
  type        = map(map(string))
  default     = {}
}

variable "frontend_cpu" {
  description = "vCPU for the frontend app"
  type        = number
  default     = 0.25
}

variable "frontend_memory" {
  description = "Memory for the frontend app"
  type        = string
  default     = "0.5Gi"
}

variable "backend_cpu" {
  description = "vCPU for the backend app"
  type        = number
  default     = 0.25
}

variable "backend_memory" {
  description = "Memory for the backend app"
  type        = string
  default     = "0.5Gi"
}

locals {
  project = "bfs"

  registry_host   = "ghcr.io"
  registry_prefix = "ly-schneider/bless2n-food-system"
  registry_user   = "ly-schneider"

  frontend_image = var.frontend_digest != "" ? "${local.registry_host}/${local.registry_prefix}/frontend@${var.frontend_digest}" : "${local.registry_host}/${local.registry_prefix}/frontend:${var.image_tag}"
  backend_image  = var.backend_digest != "" ? "${local.registry_host}/${local.registry_prefix}/backend@${var.backend_digest}" : "${local.registry_host}/${local.registry_prefix}/backend:${var.image_tag}"

  registries = [{
    server               = local.registry_host
    username             = local.registry_user
    password_secret_name = "ghcr-token"
  }]
}

output "location" {
  value = "northeurope"
}

output "tags" {
  value = {
    project    = local.project
    managed_by = "terraform"
    env        = var.env
  }
}

output "config" {
  value = {
    rg_name                  = "${local.project}-${var.env}-rg"
    env_name                 = "${local.project}-${var.env}-env"
    law_name                 = "${local.project}-logs-workspace"
    appi_name                = "${local.project}-${var.env}-insights"
    enable_app_insights      = false
    retention_days           = 30
    cosmos_name              = "${local.project}-${var.env}-cosmos"
    database_throughput      = 400
    key_vault_name           = "${local.project}-${var.env}-keyvault"
    enable_security_features = true

    apps = {
      "frontend-${var.env}" = {
        port                           = 3000
        cpu                            = var.frontend_cpu
        memory                         = var.frontend_memory
        min_replicas                   = 1
        max_replicas                   = 10
        health_check_path              = "/health"
        liveness_path                  = "/health"
        liveness_interval_seconds      = 30
        liveness_initial_delay_seconds = 20
        image                          = local.frontend_image
        revision_suffix                = var.revision_suffix
        registries                     = local.registries
        secrets = merge(
          lookup(var.app_secrets, "frontend-${var.env}", {}),
          { ghcr-token = var.ghcr_token }
        )
        environment_variables = {
          NODE_ENV                      = "production"
          LOG_LEVEL                     = "info"
          NEXT_PUBLIC_POS_PIN           = "0000"
          NEXT_PUBLIC_POS_IDLE_TIMEOUT  = "300000"
          NEXT_PUBLIC_GA_MEASUREMENT_ID = "G-9W8S03MJEM"
        }
        key_vault_secrets = merge(
          lookup(var.app_secrets, "frontend-${var.env}", {}),
          {
            "BACKEND_INTERNAL_URL"               = "backend-internal-url"
            "NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY" = "stripe-publishable-key"
            "NEXT_PUBLIC_GOOGLE_CLIENT_ID"       = "google-client-id"
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

      "backend-${var.env}" = {
        port                           = 8080
        cpu                            = var.backend_cpu
        memory                         = var.backend_memory
        min_replicas                   = 1
        max_replicas                   = 10
        health_check_path              = "/health"
        liveness_path                  = "/ping"
        liveness_interval_seconds      = 60
        liveness_initial_delay_seconds = 30
        image                          = local.backend_image
        revision_suffix                = var.revision_suffix
        registries                     = local.registries
        secrets = merge(
          lookup(var.app_secrets, "backend-${var.env}", {}),
          { ghcr-token = var.ghcr_token }
        )
        environment_variables = {
          APP_ENV                    = var.env
          APP_PORT                   = "8080"
          LOG_LEVEL                  = "info"
          LOG_DEVELOPMENT            = "false"
          SECURITY_ENABLE_HSTS       = "true"
          SECURITY_ENABLE_CSP        = "true"
          PLUNK_FROM_NAME            = "BlessThun Food"
          PLUNK_FROM_EMAIL           = ""
          PLUNK_REPLY_TO             = ""
          MONGO_DATABASE             = "bless2n_food_system"
          STATION_QR_MAX_AGE_SECONDS = "86400"
        }
        key_vault_secrets = merge(
          lookup(var.app_secrets, "backend-${var.env}", {}),
          {
            "MONGO_URI"                = "mongo-uri"
            "JWT_PRIV_PEM"             = "jwt-priv-pem"
            "JWT_PUB_PEM"              = "jwt-pub-pem"
            "STATION_QR_SECRET"        = "station-qr-secret"
            "GOOGLE_CLIENT_SECRET"     = "google-client-secret"
            "GOOGLE_CLIENT_ID"         = "google-client-id"
            "STRIPE_SECRET_KEY"        = "stripe-secret-key"
            "STRIPE_WEBHOOK_SECRET"    = "stripe-webhook-secret"
            "PLUNK_API_KEY"            = "plunk-api-key"
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
