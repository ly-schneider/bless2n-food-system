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

variable "config" {
  description = "Environment-specific configuration"
  type = object({
    rg_name                  = string
    vnet_name                = string
    subnet_name              = string
    vnet_cidr                = string
    subnet_cidr              = string
    delegate_aca_subnet      = optional(bool, false)
    pe_subnet_name           = optional(string, "private-endpoints-subnet")
    pe_subnet_cidr           = optional(string, "10.1.8.0/24")
    env_name                 = string
    law_name                 = string
    appi_name                = string
    enable_app_insights      = bool
    retention_days           = number
    cosmos_name              = string
    database_throughput      = number
    enable_security_features = optional(bool, true)
    key_vault_name           = optional(string)
    enable_private_endpoint  = optional(bool, false)
    allowed_ip_ranges        = optional(list(string), [])
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
}

module "net" {
  source                        = "../network"
  resource_group_name           = module.rg.name
  location                      = var.location
  vnet_name                     = var.config.vnet_name
  vnet_cidr                     = var.config.vnet_cidr
  subnet_name                   = var.config.subnet_name
  subnet_cidr                   = var.config.subnet_cidr
  private_endpoints_subnet_name = try(var.config.pe_subnet_name, "private-endpoints-subnet")
  private_endpoints_subnet_cidr = try(var.config.pe_subnet_cidr, "10.1.8.0/24")
  delegate_containerapps_subnet = try(var.config.delegate_aca_subnet, false)
  tags                          = var.tags
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
  subnet_id                  = module.net.subnet_id
  logs_destination           = "azure-monitor"
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

module "cosmos" {
  source                     = "../cosmos_mongo"
  name                       = var.config.cosmos_name
  location                   = var.location
  resource_group_name        = module.rg.name
  create_database            = true
  database_name              = "appdb"
  database_throughput        = var.config.database_throughput
  subnet_id                  = module.net.subnet_id
  private_endpoint_subnet_id = module.net.private_endpoints_subnet_id
  vnet_id                    = module.net.vnet_id
  allowed_ip_ranges          = []
  enable_private_endpoint    = try(var.config.enable_private_endpoint, false)
  log_analytics_workspace_id = module.obs.log_analytics_id
  tags                       = var.tags
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
  user_assigned_identity_ids     = [azurerm_user_assigned_identity.aca_uami.id]
  log_analytics_workspace_id     = module.obs.log_analytics_id
  environment_variables          = merge(local.environment_variables, each.value.environment_variables)
  secrets                        = each.value.secrets
  key_vault_secrets              = each.value.key_vault_secrets
  revision_suffix                = try(each.value.revision_suffix, null)
  key_vault_secret_refs = merge(
    try(each.value.key_vault_secret_refs, {}),
    var.config.enable_security_features && length(module.security) > 0 ? {
      for env_var, secret_name in try(each.value.key_vault_secrets, {}) :
      lower(replace(env_var, "_", "-")) => contains(keys(module.security[0].key_vault_secret_ids), secret_name) ?
      module.security[0].key_vault_secret_ids[secret_name] :
      format("%ssecrets/%s", module.security[0].key_vault_uri, secret_name)
    } : {}
  )
  registries              = each.value.registries
  http_scale_rule         = each.value.http_scale_rule
  cpu_scale_rule          = each.value.cpu_scale_rule
  memory_scale_rule       = each.value.memory_scale_rule
  azure_queue_scale_rules = each.value.azure_queue_scale_rules
  custom_scale_rules      = each.value.custom_scale_rules
  tags                    = merge(var.tags, { app = each.key })
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
  user_assigned_identity_ids     = [azurerm_user_assigned_identity.aca_uami.id]
  log_analytics_workspace_id     = module.obs.log_analytics_id
  environment_variables = merge(
    local.environment_variables,
    each.value.environment_variables
  )
  secrets           = each.value.secrets
  key_vault_secrets = each.value.key_vault_secrets
  revision_suffix   = try(each.value.revision_suffix, null)
  key_vault_secret_refs = merge(
    try(each.value.key_vault_secret_refs, {}),
    var.config.enable_security_features && length(module.security) > 0 ? {
      for env_var, secret_name in try(each.value.key_vault_secrets, {}) :
      lower(replace(env_var, "_", "-")) => contains(keys(module.security[0].key_vault_secret_ids), secret_name) ?
      module.security[0].key_vault_secret_ids[secret_name] :
      format("%ssecrets/%s", module.security[0].key_vault_uri, secret_name)
    } : {}
  )
  registries              = each.value.registries
  http_scale_rule         = each.value.http_scale_rule
  cpu_scale_rule          = each.value.cpu_scale_rule
  memory_scale_rule       = each.value.memory_scale_rule
  azure_queue_scale_rules = each.value.azure_queue_scale_rules
  custom_scale_rules      = each.value.custom_scale_rules
  tags                    = merge(var.tags, { app = each.key })
}

data "azurerm_client_config" "current" {}

resource "azurerm_user_assigned_identity" "aca_uami" {
  name                = "${var.environment}-aca-uami"
  resource_group_name = module.rg.name
  location            = var.location
  tags                = var.tags
}

module "security" {
  count = var.config.enable_security_features ? 1 : 0

  source                     = "../security"
  name_prefix                = var.environment
  location                   = var.location
  resource_group_name        = module.rg.name
  subnet_id                  = module.net.subnet_id
  enable_key_vault           = true
  key_vault_name             = var.config.key_vault_name != null ? var.config.key_vault_name : "${var.config.cosmos_name}-kv"
  allowed_ip_ranges          = var.config.allowed_ip_ranges != null ? var.config.allowed_ip_ranges : []
  key_vault_admins           = [data.azurerm_client_config.current.object_id]
  uami_principal_id          = azurerm_user_assigned_identity.aca_uami.principal_id
  cosmos_connection_string   = module.cosmos.connection_string
  log_analytics_workspace_id = module.obs.log_analytics_id
  tags                       = var.tags
}
