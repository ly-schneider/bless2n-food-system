variable "environment" {
  description = "Environment name"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}

variable "dns" {
  description = "DNS configuration from defaults module"
  type = object({
    enabled         = bool
    base_domain     = string
    domain_prefix   = string
    frontend_domain = string
    backend_domain  = string
    docs_domain     = string
    frontend_url    = string
    backend_url     = string
  })
}

variable "config" {
  description = "Environment-specific configuration"
  type = object({
    rg_name             = string
    env_name            = string
    law_name            = string
    appi_name           = string
    enable_app_insights = bool
    retention_days      = number
    key_vault_name      = optional(string)
    allowed_ip_ranges   = optional(list(string), [])
    apps = map(object({
      port                           = number
      image                          = string
      cpu                            = number
      memory                         = string
      min_replicas                   = number
      max_replicas                   = number
      revision_suffix                = optional(string)
      external_ingress               = optional(bool, true)
      allow_insecure_connections     = optional(bool, false)
      health_check_path              = optional(string)
      liveness_path                  = optional(string)
      liveness_interval_seconds      = optional(number)
      liveness_initial_delay_seconds = optional(number)
      environment_variables          = optional(map(string), {})
      secrets                        = optional(map(string), {})
      key_vault_secrets              = optional(map(string), {})
      key_vault_secret_refs          = optional(map(string), {})
      registries = optional(list(object({
        server               = string
        username             = optional(string)
        password_secret_name = optional(string)
        identity             = optional(string)
      })), [])
      http_scale_rule = optional(object({
        name                = string
        concurrent_requests = number
      }))
      cpu_scale_rule = optional(object({
        name           = string
        cpu_percentage = number
      }))
      memory_scale_rule = optional(object({
        name              = string
        memory_percentage = number
      }))
      azure_queue_scale_rules = optional(list(object({
        name              = string
        queue_name        = string
        queue_length      = number
        secret_name       = string
        trigger_parameter = string
      })), [])
      custom_scale_rules = optional(list(object({
        name             = string
        custom_rule_type = string
        metadata         = map(string)
        authentication = optional(object({
          secret_name       = string
          trigger_parameter = string
        }))
      })), [])
    }))
  })
}

locals {
  environment_variables = var.config.enable_app_insights ? {
    APPINSIGHTS_CONNECTION_STRING = coalesce(module.obs.app_insights_connection_string, "")
  } : {}
  key_vault_name   = var.config.key_vault_name != null ? var.config.key_vault_name : "${var.environment}-kv"
  key_vault_uri    = "https://${local.key_vault_name}.vault.azure.net"
  uami_name        = "${var.environment}-aca-uami"
  uami_resource_id = "/subscriptions/${data.azurerm_client_config.current.subscription_id}/resourceGroups/${var.config.rg_name}/providers/Microsoft.ManagedIdentity/userAssignedIdentities/${local.uami_name}"
}

module "rg" {
  source   = "../rg"
  name     = var.config.rg_name
  location = var.location
  tags     = var.tags
}

locals {
  backend_apps  = { for k, v in var.config.apps : k => v if can(regex("^backend", k)) }
  frontend_apps = { for k, v in var.config.apps : k => v if can(regex("^frontend", k)) }
  docs_apps     = { for k, v in var.config.apps : k => v if can(regex("^docs", k)) }
}

module "obs" {
  source              = "../observability"
  resource_group_name = module.rg.name
  location            = var.location
  law_name            = var.config.law_name
  appi_name           = var.config.appi_name
  enable_app_insights = var.config.enable_app_insights
  retention_days      = var.config.retention_days
  tags                = var.tags
}

module "aca_env" {
  source                     = "../containerapps_env"
  name                       = var.config.env_name
  location                   = var.location
  resource_group_name        = module.rg.name
  log_analytics_workspace_id = module.obs.log_analytics_id
  tags                       = var.tags
}

module "env_diag" {
  source                     = "../diagnostic_setting"
  name                       = "${var.config.env_name}-diag"
  target_resource_id         = module.aca_env.id
  log_analytics_workspace_id = module.obs.log_analytics_id
  category_groups            = []
  enable_metrics             = true
}

module "apps_backend" {
  for_each = local.backend_apps
  source   = "../containerapp"

  name                           = each.key
  resource_group_name            = module.rg.name
  environment_id                 = module.aca_env.id
  image                          = each.value.image
  target_port                    = each.value.port
  health_check_path              = lookup(each.value, "health_check_path", "/health")
  liveness_path                  = lookup(each.value, "liveness_path", "/health")
  liveness_interval_seconds      = lookup(each.value, "liveness_interval_seconds", 60)
  liveness_initial_delay_seconds = lookup(each.value, "liveness_initial_delay_seconds", 20)
  external_ingress               = try(each.value.external_ingress, true)
  allow_insecure_connections     = try(each.value.allow_insecure_connections, false)
  cpu                            = each.value.cpu
  memory                         = each.value.memory
  min_replicas                   = each.value.min_replicas
  max_replicas                   = each.value.max_replicas
  enable_system_identity         = false
  user_assigned_identity_ids     = [local.uami_resource_id]
  log_analytics_workspace_id     = module.obs.log_analytics_id
  environment_variables = merge(local.environment_variables, each.value.environment_variables, {
    AZURE_STORAGE_CONTAINER     = module.blob_storage.container_name
    AZURE_STORAGE_BLOB_ENDPOINT = module.blob_storage.blob_endpoint
  })
  secrets = merge(each.value.secrets, {
    azure-storage-account-name = module.blob_storage.storage_account_name
    azure-storage-account-key  = module.blob_storage.primary_access_key
  })
  key_vault_secrets = merge(each.value.key_vault_secrets, {
    "AZURE_STORAGE_ACCOUNT_NAME" = "azure-storage-account-name"
    "AZURE_STORAGE_ACCOUNT_KEY"  = "azure-storage-account-key"
  })
  revision_suffix = try(each.value.revision_suffix, null)
  key_vault_secret_refs = {
    for secret_name in distinct(values(try(each.value.key_vault_secrets, {}))) :
    secret_name => "${local.key_vault_uri}/secrets/${secret_name}"
  }
  registries              = each.value.registries
  http_scale_rule         = each.value.http_scale_rule
  cpu_scale_rule          = each.value.cpu_scale_rule
  memory_scale_rule       = each.value.memory_scale_rule
  azure_queue_scale_rules = each.value.azure_queue_scale_rules
  custom_scale_rules      = each.value.custom_scale_rules
  tags                    = merge(var.tags, { app = each.key })

  depends_on = [
    module.security,
    module.blob_storage,
    azurerm_user_assigned_identity.aca_uami,
  ]
}

module "apps_frontend" {
  for_each = local.frontend_apps
  source   = "../containerapp"

  name                           = each.key
  resource_group_name            = module.rg.name
  environment_id                 = module.aca_env.id
  image                          = each.value.image
  target_port                    = each.value.port
  health_check_path              = lookup(each.value, "health_check_path", "/api/health")
  liveness_path                  = lookup(each.value, "liveness_path", "/api/health")
  liveness_interval_seconds      = lookup(each.value, "liveness_interval_seconds", 30)
  liveness_initial_delay_seconds = lookup(each.value, "liveness_initial_delay_seconds", 20)
  external_ingress               = try(each.value.external_ingress, true)
  allow_insecure_connections     = try(each.value.allow_insecure_connections, false)
  cpu                            = each.value.cpu
  memory                         = each.value.memory
  min_replicas                   = each.value.min_replicas
  max_replicas                   = each.value.max_replicas
  enable_system_identity         = false
  user_assigned_identity_ids     = [local.uami_resource_id]
  log_analytics_workspace_id     = module.obs.log_analytics_id
  environment_variables = merge(
    local.environment_variables,
    each.value.environment_variables,
    {
      BACKEND_INTERNAL_URL = "https://${module.apps_backend["backend-${var.environment}"].fqdn}"
    }
  )
  secrets           = each.value.secrets
  key_vault_secrets = each.value.key_vault_secrets
  revision_suffix   = try(each.value.revision_suffix, null)
  key_vault_secret_refs = {
    for secret_name in distinct(values(try(each.value.key_vault_secrets, {}))) :
    secret_name => "${local.key_vault_uri}/secrets/${secret_name}"
  }
  registries              = each.value.registries
  http_scale_rule         = each.value.http_scale_rule
  cpu_scale_rule          = each.value.cpu_scale_rule
  memory_scale_rule       = each.value.memory_scale_rule
  azure_queue_scale_rules = each.value.azure_queue_scale_rules
  custom_scale_rules      = each.value.custom_scale_rules
  tags                    = merge(var.tags, { app = each.key })

  depends_on = [
    module.security,
    module.apps_backend,
    azurerm_user_assigned_identity.aca_uami,
  ]
}

module "apps_docs" {
  for_each = local.docs_apps
  source   = "../containerapp"

  name                           = each.key
  resource_group_name            = module.rg.name
  environment_id                 = module.aca_env.id
  image                          = each.value.image
  target_port                    = each.value.port
  health_check_path              = lookup(each.value, "health_check_path", "/api/health")
  liveness_path                  = lookup(each.value, "liveness_path", "/api/health")
  liveness_interval_seconds      = lookup(each.value, "liveness_interval_seconds", 30)
  liveness_initial_delay_seconds = lookup(each.value, "liveness_initial_delay_seconds", 20)
  external_ingress               = try(each.value.external_ingress, true)
  allow_insecure_connections     = try(each.value.allow_insecure_connections, false)
  cpu                            = each.value.cpu
  memory                         = each.value.memory
  min_replicas                   = each.value.min_replicas
  max_replicas                   = each.value.max_replicas
  enable_system_identity         = false
  user_assigned_identity_ids     = [local.uami_resource_id]
  log_analytics_workspace_id     = module.obs.log_analytics_id
  environment_variables = merge(
    local.environment_variables,
    each.value.environment_variables
  )
  secrets           = each.value.secrets
  key_vault_secrets = each.value.key_vault_secrets
  revision_suffix   = try(each.value.revision_suffix, null)
  key_vault_secret_refs = {
    for secret_name in distinct(values(try(each.value.key_vault_secrets, {}))) :
    secret_name => "${local.key_vault_uri}/secrets/${secret_name}"
  }
  registries              = each.value.registries
  http_scale_rule         = each.value.http_scale_rule
  cpu_scale_rule          = each.value.cpu_scale_rule
  memory_scale_rule       = each.value.memory_scale_rule
  azure_queue_scale_rules = each.value.azure_queue_scale_rules
  custom_scale_rules      = each.value.custom_scale_rules
  tags                    = merge(var.tags, { app = each.key })

  depends_on = [
    module.security,
    azurerm_user_assigned_identity.aca_uami,
  ]
}

data "azurerm_client_config" "current" {}

resource "azurerm_user_assigned_identity" "aca_uami" {
  name                = "${var.environment}-aca-uami"
  resource_group_name = module.rg.name
  location            = var.location
  tags                = var.tags
}

module "blob_storage" {
  source              = "../blob_storage"
  name                = "bfs${var.environment}images"
  resource_group_name = module.rg.name
  location            = var.location
  container_name      = "product-images"
  tags                = var.tags
}

module "security" {
  source              = "../security"
  name_prefix         = var.environment
  location            = var.location
  resource_group_name = module.rg.name
  key_vault_name      = local.key_vault_name
  tags                = var.tags
}

# --- DNS ---

data "azurerm_dns_zone" "food" {
  count               = var.dns.enabled ? 1 : 0
  name                = var.dns.base_domain
  resource_group_name = "bfs-dns-rg"
}

locals {
  dns_enabled      = var.dns.enabled
  is_apex          = var.dns.domain_prefix == ""
  dns_zone_name    = local.dns_enabled ? data.azurerm_dns_zone.food[0].name : ""
  dns_zone_rg      = local.dns_enabled ? data.azurerm_dns_zone.food[0].resource_group_name : ""
  backend_record   = local.is_apex ? "api" : "api.${var.dns.domain_prefix}"
  docs_record      = local.is_apex ? "docs" : "docs.${var.dns.domain_prefix}"
  asuid_frontend   = local.is_apex ? "asuid" : "asuid.${var.dns.domain_prefix}"
  asuid_backend    = local.is_apex ? "asuid.api" : "asuid.api.${var.dns.domain_prefix}"
  asuid_docs       = local.is_apex ? "asuid.docs" : "asuid.docs.${var.dns.domain_prefix}"
  verification_id  = module.aca_env.custom_domain_verification_id
  frontend_app_key = "frontend-${var.environment}"
  backend_app_key  = "backend-${var.environment}"
  docs_app_key     = "docs-${var.environment}"
}

resource "azurerm_dns_a_record" "frontend" {
  count               = local.dns_enabled && local.is_apex ? 1 : 0
  name                = "@"
  zone_name           = local.dns_zone_name
  resource_group_name = local.dns_zone_rg
  ttl                 = 300
  records             = [module.aca_env.static_ip_address]
}

resource "azurerm_dns_cname_record" "frontend" {
  count               = local.dns_enabled && !local.is_apex ? 1 : 0
  name                = var.dns.domain_prefix
  zone_name           = local.dns_zone_name
  resource_group_name = local.dns_zone_rg
  ttl                 = 300
  record              = module.apps_frontend[local.frontend_app_key].fqdn
}

resource "azurerm_dns_cname_record" "backend" {
  count               = local.dns_enabled ? 1 : 0
  name                = local.backend_record
  zone_name           = local.dns_zone_name
  resource_group_name = local.dns_zone_rg
  ttl                 = 300
  record              = module.apps_backend[local.backend_app_key].fqdn
}

resource "azurerm_dns_cname_record" "docs" {
  count               = local.dns_enabled ? 1 : 0
  name                = local.docs_record
  zone_name           = local.dns_zone_name
  resource_group_name = local.dns_zone_rg
  ttl                 = 300
  record              = module.apps_docs[local.docs_app_key].fqdn
}

resource "azurerm_dns_txt_record" "asuid_frontend" {
  count               = local.dns_enabled ? 1 : 0
  name                = local.asuid_frontend
  zone_name           = local.dns_zone_name
  resource_group_name = local.dns_zone_rg
  ttl                 = 300

  record {
    value = local.verification_id
  }
}

resource "azurerm_dns_txt_record" "asuid_backend" {
  count               = local.dns_enabled ? 1 : 0
  name                = local.asuid_backend
  zone_name           = local.dns_zone_name
  resource_group_name = local.dns_zone_rg
  ttl                 = 300

  record {
    value = local.verification_id
  }
}

resource "azurerm_dns_txt_record" "asuid_docs" {
  count               = local.dns_enabled ? 1 : 0
  name                = local.asuid_docs
  zone_name           = local.dns_zone_name
  resource_group_name = local.dns_zone_rg
  ttl                 = 300

  record {
    value = local.verification_id
  }
}

resource "azurerm_container_app_custom_domain" "frontend" {
  count                    = local.dns_enabled ? 1 : 0
  name                     = var.dns.frontend_domain
  container_app_id         = module.apps_frontend[local.frontend_app_key].id
  certificate_binding_type = "Auto"

  lifecycle {
    ignore_changes = [
      certificate_binding_type,
      container_app_environment_certificate_id
    ]
  }

  depends_on = [
    azurerm_dns_a_record.frontend,
    azurerm_dns_cname_record.frontend,
    azurerm_dns_txt_record.asuid_frontend,
  ]
}

resource "azurerm_container_app_custom_domain" "backend" {
  count                    = local.dns_enabled ? 1 : 0
  name                     = var.dns.backend_domain
  container_app_id         = module.apps_backend[local.backend_app_key].id
  certificate_binding_type = "Auto"

  lifecycle {
    ignore_changes = [
      certificate_binding_type,
      container_app_environment_certificate_id
    ]
  }

  depends_on = [
    azurerm_dns_cname_record.backend,
    azurerm_dns_txt_record.asuid_backend,
  ]
}

resource "azurerm_container_app_custom_domain" "docs" {
  count                    = local.dns_enabled ? 1 : 0
  name                     = var.dns.docs_domain
  container_app_id         = module.apps_docs[local.docs_app_key].id
  certificate_binding_type = "Auto"

  lifecycle {
    ignore_changes = [
      certificate_binding_type,
      container_app_environment_certificate_id
    ]
  }

  depends_on = [
    azurerm_dns_cname_record.docs,
    azurerm_dns_txt_record.asuid_docs,
  ]
}
