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

variable "docs_digest" {
  description = "Docs image digest (optional, preferred over tag)"
  type        = string
  default     = ""
}

variable "domain_prefix" {
  description = "Subdomain prefix (empty for production apex)"
  type        = string
}

variable "enable_dns" {
  description = "Manage DNS records and custom domain bindings via Azure DNS"
  type        = bool
  default     = false
}

locals {
  project = "bfs"

  base_domain     = "food.blessthun.ch"
  frontend_domain = var.domain_prefix != "" ? "${var.domain_prefix}.${local.base_domain}" : local.base_domain
  backend_domain  = var.domain_prefix != "" ? "api.${var.domain_prefix}.${local.base_domain}" : "api.${local.base_domain}"
  docs_domain     = var.domain_prefix != "" ? "docs.${var.domain_prefix}.${local.base_domain}" : "docs.${local.base_domain}"
  frontend_url    = "https://${local.frontend_domain}"
  backend_url     = "https://${local.backend_domain}"

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

output "dns" {
  value = {
    enabled         = var.enable_dns
    base_domain     = local.base_domain
    domain_prefix   = var.domain_prefix
    frontend_domain = local.frontend_domain
    backend_domain  = local.backend_domain
    docs_domain     = local.docs_domain
    frontend_url    = local.frontend_url
    backend_url     = local.backend_url
  }
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
        environment_variables = {
          NODE_ENV                      = "production"
          LOG_LEVEL                     = "info"
          APP_VERSION                   = var.image_tag
          NEXT_PUBLIC_POS_PIN           = "0000"
          NEXT_PUBLIC_POS_IDLE_TIMEOUT  = "300000"
          NEXT_PUBLIC_GA_MEASUREMENT_ID = "G-9W8S03MJEM"
          NEXT_PUBLIC_API_BASE_URL      = local.backend_url
          NEXT_PUBLIC_APP_URL           = local.frontend_url
          BETTER_AUTH_URL               = local.frontend_url
        }
        key_vault_secrets = {
          "BETTER_AUTH_SECRET"   = "better-auth-secret"
          "DATABASE_URL"         = "database-url"
          "GOOGLE_CLIENT_ID"     = "google-client-id"
          "GOOGLE_CLIENT_SECRET" = "google-client-secret"
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
        environment_variables = {
          NODE_ENV    = "production"
          APP_VERSION = var.image_tag
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
        environment_variables = {
          APP_ENV                  = var.env
          APP_PORT                 = "8080"
          APP_VERSION              = var.image_tag
          LOG_LEVEL                = "info"
          LOG_DEVELOPMENT          = "false"
          SECURITY_ENABLE_HSTS     = "true"
          SECURITY_ENABLE_CSP      = "true"
          PLUNK_FROM_NAME          = "BlessThun Food"
          PLUNK_FROM_EMAIL         = ""
          PLUNK_REPLY_TO           = ""
          BETTER_AUTH_URL          = local.frontend_url
          PUBLIC_BASE_URL          = local.frontend_url
          SECURITY_TRUSTED_ORIGINS = local.frontend_url
        }
        key_vault_secrets = {
          "DATABASE_URL"           = "database-url"
          "PAYREXX_INSTANCE"       = "payrexx-instance"
          "PAYREXX_API_SECRET"     = "payrexx-api-secret"
          "PAYREXX_WEBHOOK_SECRET" = "payrexx-webhook-secret"
          "PLUNK_API_KEY"          = "plunk-api-key"
          "ELVANTO_API_KEY"        = "elvanto-api-key"
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
