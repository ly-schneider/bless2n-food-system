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
    # Whether to delegate the Container Apps subnet (Consumption needs delegation; Workload Profiles requires non-delegated).
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
    enable_alerts            = bool
    requests_5xx_threshold   = number
    enable_security_features = optional(bool, true)
    key_vault_name           = optional(string)
    enable_private_endpoint  = optional(bool, false)
    allowed_ip_ranges        = optional(list(string), [])
    budget_amount            = optional(number)
    budget_start_date        = optional(string, "2025-01-01T00:00:00Z")
    apps = map(object({
      port            = number
      image           = string
      cpu             = number
      memory          = string
      min_replicas    = number
      max_replicas    = number
      revision_suffix = optional(string)
      # Probes
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
  # Split apps by convention to allow ordering: backends first, then frontends
  # This enables wiring frontend -> backend using the backend's stable FQDN via data.azurerm_container_app
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
  depends_on                    = [module.tfc_rbac]
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

module "tfc_rbac" {
  source = "../rbac_tfc"

  # Scope all assignments to the environment resource group
  target_rg_id = module.rg.id

  # Grant least-privilege roles required by this stack
  network_scopes             = [module.rg.id]
  private_dns_zone_scopes    = try(var.config.enable_private_endpoint, false) ? [module.rg.id] : []
  managed_identity_scopes    = [module.rg.id]
  uaa_scopes                 = [module.rg.id]
  cosmos_account_scopes      = [module.rg.id]
  grant_cosmos_account_contributor = true
  grant_cosmos_keys_reader         = false
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
  category_groups = []
  enable_metrics  = true
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

  # Ensure RBAC assignments (e.g., Cosmos roles) are in place before reading keys/connection string
  depends_on = [module.tfc_rbac]
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
  registries = each.value.registries
  http_scale_rule         = each.value.http_scale_rule
  cpu_scale_rule          = each.value.cpu_scale_rule
  memory_scale_rule       = each.value.memory_scale_rule
  azure_queue_scale_rules = each.value.azure_queue_scale_rules
  custom_scale_rules      = each.value.custom_scale_rules
  tags                    = merge(var.tags, { app = each.key })
}

# Backend app FQDNs are available from the module output; no data source needed

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
  # Resolve Key Vault secret IDs inside the stack to avoid referencing module outputs from the caller
  # Prefer module.security-provided versionless IDs; otherwise, construct proper Key Vault URI format.
  key_vault_secret_refs = merge(
    try(each.value.key_vault_secret_refs, {}),
    var.config.enable_security_features && length(module.security) > 0 ? {
      for env_var, secret_name in try(each.value.key_vault_secrets, {}) :
      lower(replace(env_var, "_", "-")) => contains(keys(module.security[0].key_vault_secret_ids), secret_name) ?
      module.security[0].key_vault_secret_ids[secret_name] :
      format("%ssecrets/%s", module.security[0].key_vault_uri, secret_name)
    } : {}
  )
  registries = each.value.registries
  http_scale_rule         = each.value.http_scale_rule
  cpu_scale_rule          = each.value.cpu_scale_rule
  memory_scale_rule       = each.value.memory_scale_rule
  azure_queue_scale_rules = each.value.azure_queue_scale_rules
  custom_scale_rules      = each.value.custom_scale_rules
  tags                    = merge(var.tags, { app = each.key })
}

module "alerts" {
  count = var.config.enable_alerts ? 1 : 0

  source              = "../alerts"
  name                = "${var.environment}-alerts"
  short_name          = var.environment
  resource_group_name = module.rg.name
  email_receivers     = var.alert_emails
  action_group_id     = module.ag.id
  container_app_ids = merge(
    { for k, m in module.apps_backend : k => m.id },
    { for k, m in module.apps_frontend : k => m.id }
  )
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

module "rg_budget" {
  count  = try(var.config.budget_amount, null) != null ? 1 : 0
  source = "../budget"

  name              = "${var.environment}-rg-budget"
  resource_group_id = module.rg.id
  amount            = var.config.budget_amount
  start_date        = try(var.config.budget_start_date, "2025-01-01T00:00:00Z")
  action_group_id   = module.ag.id
  tags              = var.tags
}

module "ag" {
  source = "../action_group"

  name                = "bfs-${var.environment}-ag"
  short_name          = var.environment
  resource_group_name = module.rg.name
  email_receivers     = var.alert_emails
  tags                = var.tags
}
