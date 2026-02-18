# Centralized configuration module
# Takes environment name and optional per-env inputs, outputs complete ready-to-use config

variable "env" {
  description = "Environment name (staging, production)"
  type        = string
}

variable "image_tag" {
  description = "Semantic version tag for Docker images"
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

variable "frontend_custom_domains" {
  description = "Custom domains for the frontend app"
  type        = list(string)
  default     = []
}

variable "backend_custom_domains" {
  description = "Custom domains for the backend app"
  type        = list(string)
  default     = []
}

variable "docs_digest" {
  description = "Docs image digest (optional, preferred over tag)"
  type        = string
  default     = ""
}

variable "docs_custom_domains" {
  description = "Custom domains for the docs app"
  type        = list(string)
  default     = []
}

locals {
  project = "bfs"

  registry_host   = "ghcr.io"
  registry_prefix = "ly-schneider/bless2n-food-system"
  registry_user   = "ly-schneider"

  frontend_image = var.frontend_digest != "" ? "${local.registry_host}/${local.registry_prefix}/frontend@${var.frontend_digest}" : "${local.registry_host}/${local.registry_prefix}/frontend:${var.image_tag}"
  backend_image  = var.backend_digest != "" ? "${local.registry_host}/${local.registry_prefix}/backend@${var.backend_digest}" : "${local.registry_host}/${local.registry_prefix}/backend:${var.image_tag}"
  docs_image     = var.docs_digest != "" ? "${local.registry_host}/${local.registry_prefix}/docs@${var.docs_digest}" : "${local.registry_host}/${local.registry_prefix}/docs:${var.image_tag}"

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
    project = local.project
    env     = var.env
  }
}

output "config" {
  value = {
    rg_name             = "${local.project}-${var.env}-rg"
    env_name            = "${local.project}-${var.env}-env"
    law_name            = "${local.project}-logs-workspace"
    appi_name           = "${local.project}-${var.env}-insights"
    enable_app_insights = false
    retention_days      = 30
    key_vault_name      = "${local.project}-${var.env}-keyvault"

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
        secrets                        = { ghcr-token = var.ghcr_token }
        custom_domains                 = var.frontend_custom_domains
        environment_variables = {
          NODE_ENV                      = "production"
          LOG_LEVEL                     = "info"
          NEXT_PUBLIC_POS_PIN           = "0000"
          NEXT_PUBLIC_POS_IDLE_TIMEOUT  = "300000"
          NEXT_PUBLIC_GA_MEASUREMENT_ID = "G-9W8S03MJEM"
        }
        key_vault_secrets = {
          "BACKEND_INTERNAL_URL"     = "backend-internal-url"
          "NEXT_PUBLIC_API_BASE_URL" = "next-public-api-base-url"
          "NEXT_PUBLIC_APP_URL"      = "next-public-app-url"
          "BETTER_AUTH_SECRET"       = "better-auth-secret"
          "BETTER_AUTH_URL"          = "better-auth-url"
          "DATABASE_URL"             = "database-url"
          "GOOGLE_CLIENT_ID"         = "google-client-id"
          "GOOGLE_CLIENT_SECRET"     = "google-client-secret"
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

      "docs-${var.env}" = {
        port                           = 3000
        cpu                            = 0.25
        memory                         = "0.5Gi"
        min_replicas                   = 0
        max_replicas                   = 3
        health_check_path              = "/api/health"
        liveness_path                  = "/api/health"
        liveness_interval_seconds      = 30
        liveness_initial_delay_seconds = 20
        image                          = local.docs_image
        revision_suffix                = var.revision_suffix
        registries                     = local.registries
        secrets                        = { ghcr-token = var.ghcr_token }
        custom_domains                 = var.docs_custom_domains
        environment_variables = {
          NODE_ENV = "production"
        }
        key_vault_secrets = {}
        http_scale_rule = {
          name                = "docs-http-scale"
          concurrent_requests = 10
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
        secrets                        = { ghcr-token = var.ghcr_token }
        custom_domains                 = var.backend_custom_domains
        environment_variables = {
          APP_ENV                 = var.env
          APP_PORT                = "8080"
          LOG_LEVEL               = "info"
          LOG_DEVELOPMENT         = "false"
          SECURITY_ENABLE_HSTS    = "true"
          SECURITY_ENABLE_CSP     = "true"
          PLUNK_FROM_NAME         = "BlessThun Food"
          PLUNK_FROM_EMAIL        = ""
          PLUNK_REPLY_TO          = ""
          AZURE_STORAGE_CONTAINER = "product-images"
        }
        key_vault_secrets = {
          "DATABASE_URL"               = "database-url"
          "BETTER_AUTH_URL"            = "better-auth-url"
          "PUBLIC_BASE_URL"            = "public-base-url"
          "SECURITY_TRUSTED_ORIGINS"   = "security-trusted-origins"
          "PAYREXX_INSTANCE"           = "payrexx-instance"
          "PAYREXX_API_SECRET"         = "payrexx-api-secret"
          "PAYREXX_WEBHOOK_SECRET"     = "payrexx-webhook-secret"
          "PLUNK_API_KEY"              = "plunk-api-key"
          "AZURE_STORAGE_ACCOUNT_NAME" = "azure-storage-account-name"
          "AZURE_STORAGE_ACCOUNT_KEY"  = "azure-storage-account-key"
        }
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
