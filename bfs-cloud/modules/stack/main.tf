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

variable "alert_emails" {
  description = "Email addresses for alerts"
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
    pe_subnet_name           = optional(string, "private-endpoints-subnet")
    pe_subnet_cidr           = optional(string, "10.1.8.0/24")
    env_name                 = string
    law_name                 = string
    appi_name                = string
    enable_app_insights      = bool
    retention_days           = number
    cosmos_name              = string
    database_throughput      = number
    enable_alerts            = bool
    requests_5xx_threshold   = number
    enable_security_features = optional(bool, true)
    enable_acr               = optional(bool, true)
    acr_login_server         = optional(string)
    acr_resource_id          = optional(string)
    acr_name                 = optional(string)
    acr_sku                  = optional(string, "Basic")
    key_vault_name           = optional(string)
    allowed_ip_ranges        = optional(list(string), [])
    apps = map(object({
      port                  = number
      image                 = string
      cpu                   = number
      memory                = string
      min_replicas          = number
      max_replicas          = number
      environment_variables = optional(map(string), {})
      secrets               = optional(map(string), {})
      key_vault_secrets     = optional(map(string), {})
      key_vault_secret_refs = optional(map(string), {})
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

# Create ACR if enabled
module "acr" {
  count  = try(var.config.enable_acr, true) ? 1 : 0
  source = "../acr"

  name                = var.config.acr_name
  resource_group_name = module.rg.name
  location            = var.location
  sku                 = var.config.acr_sku
  admin_enabled       = false

  depends_on = [module.rg]
}

data "azurerm_container_registry" "acr" {
  count               = var.config.acr_login_server == null && var.config.acr_name != null && try(var.config.enable_acr, true) == false ? 1 : 0
  name                = var.config.acr_name
  resource_group_name = module.rg.name
}

locals {
  acr_login_server = try(
    module.acr[0].login_server,
    try(data.azurerm_container_registry.acr[0].login_server, var.config.acr_login_server)
  )
  acr_scope_id = try(
    module.acr[0].id,
    try(data.azurerm_container_registry.acr[0].id, var.config.acr_resource_id)
  )
  # Enable ACR pull role assignment when we can determine ACR by inputs alone.
  # This avoids making `count` depend on computed values that are unknown at plan time.
  enable_uami_acr_pull = (
    try(var.config.enable_acr, true)
    || var.config.acr_resource_id != null
    || (
      var.config.acr_login_server == null &&
      var.config.acr_name != null &&
      try(var.config.enable_acr, true) == false
    )
  )
}

module "net" {
  source              = "../network"
  resource_group_name = module.rg.name
  location            = var.location
  vnet_name           = var.config.vnet_name
  vnet_cidr           = var.config.vnet_cidr
  subnet_name         = var.config.subnet_name
  subnet_cidr         = var.config.subnet_cidr
  private_endpoints_subnet_name = try(var.config.pe_subnet_name, "private-endpoints-subnet")
  private_endpoints_subnet_cidr = try(var.config.pe_subnet_cidr, "10.1.8.0/24")
  tags                = var.tags
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
  # Container Apps Environment does not support category_group "allLogs".
  # Use metrics only for now to avoid API errors.
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
  log_analytics_workspace_id = module.obs.log_analytics_id
  tags                       = var.tags
}

module "apps" {
  for_each = var.config.apps
  source   = "../containerapp"

  name                       = each.key
  resource_group_name        = module.rg.name
  environment_id             = module.aca_env.id
  image                      = each.value.image
  target_port                = each.value.port
  health_check_path          = lookup(each.value, "health_check_path", "/health")
  external_ingress           = try(each.value.external_ingress, true)
  cpu                        = each.value.cpu
  memory                     = each.value.memory
  min_replicas               = each.value.min_replicas
  max_replicas               = each.value.max_replicas
  enable_system_identity     = false
  user_assigned_identity_ids = [azurerm_user_assigned_identity.aca_uami.id]
  log_analytics_workspace_id = module.obs.log_analytics_id
  environment_variables      = merge(local.environment_variables, each.value.environment_variables)
  secrets                    = each.value.secrets
  key_vault_secrets          = each.value.key_vault_secrets
  # Resolve Key Vault secret IDs inside the stack to avoid referencing module outputs from the caller
  key_vault_secret_refs = try(
    merge(
      each.value.key_vault_secret_refs,
      var.config.enable_security_features && length(module.security) > 0 ? {
        for s in distinct(values(try(each.value.key_vault_secrets, {}))) : s => module.security[0].key_vault_secret_ids[s]
        if contains(keys(module.security[0].key_vault_secret_ids), s)
      } : {}
    ),
    var.config.enable_security_features && length(module.security) > 0 ? {
      for s in distinct(values(try(each.value.key_vault_secrets, {}))) : s => module.security[0].key_vault_secret_ids[s]
      if contains(keys(module.security[0].key_vault_secret_ids), s)
    } : {}
  )
  registries = concat(
    local.acr_login_server != null ? [
      {
        server   = local.acr_login_server
        identity = azurerm_user_assigned_identity.aca_uami.id
      }
    ] : [],
    each.value.registries
  )
  http_scale_rule         = each.value.http_scale_rule
  cpu_scale_rule          = each.value.cpu_scale_rule
  memory_scale_rule       = each.value.memory_scale_rule
  azure_queue_scale_rules = each.value.azure_queue_scale_rules
  custom_scale_rules      = each.value.custom_scale_rules
  tags                    = merge(var.tags, { app = each.key })
}

# Grant UAMI pull access to ACR when enabled
resource "azurerm_role_assignment" "uami_acr_pull" {
  # Use a boolean derived only from input variables so `count` is known during plan.
  count                = local.enable_uami_acr_pull ? 1 : 0
  scope                = local.acr_scope_id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_user_assigned_identity.aca_uami.principal_id
}

module "alerts" {
  count = var.config.enable_alerts ? 1 : 0

  source                 = "../alerts"
  name                   = "${var.environment}-alerts"
  short_name             = var.environment
  resource_group_name    = module.rg.name
  email_receivers        = var.alert_emails
  container_app_ids      = { for k, m in module.apps : k => m.id }
  requests_5xx_threshold = var.config.requests_5xx_threshold
  tags                   = var.tags
}

data "azurerm_client_config" "current" {}

# Create User-Assigned Managed Identity for Container Apps
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
  enable_basic_monitoring    = false
  container_app_ids          = {}
  action_group_id            = null
  log_analytics_workspace_id = module.obs.log_analytics_id
  tags                       = var.tags
}
