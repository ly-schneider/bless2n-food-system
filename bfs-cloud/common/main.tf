terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
  }
}

provider "azurerm" {
  features {}
  resource_provider_registrations = "none"
}

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
    rg_name                      = string
    vnet_name                    = string
    subnet_name                  = string
    vnet_cidr                    = string
    subnet_cidr                  = string
    env_name                     = string
    law_name                     = string
    appi_name                    = string
    enable_app_insights          = bool
    retention_days               = number
    cosmos_name                  = string
    database_throughput          = number
    enable_alerts                = bool
    requests_5xx_threshold       = number
    enable_security_features     = optional(bool, true)
    key_vault_name               = optional(string)
    apps = map(object({
      port         = number
      image        = string
      cpu          = number
      memory       = string
      min_replicas = number
      max_replicas = number
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
  source   = "../modules/rg"
  name     = var.config.rg_name
  location = var.location
  tags     = var.tags
}

module "net" {
  source              = "../modules/network"
  resource_group_name = module.rg.name
  location            = var.location
  vnet_name           = var.config.vnet_name
  vnet_cidr           = var.config.vnet_cidr
  subnet_name         = var.config.subnet_name
  subnet_cidr         = var.config.subnet_cidr
  tags                = var.tags
}

module "obs" {
  source               = "../modules/observability"
  resource_group_name  = module.rg.name
  location             = var.location
  law_name             = var.config.law_name
  appi_name            = var.config.appi_name
  enable_app_insights  = var.config.enable_app_insights
  retention_days       = var.config.retention_days
  tags                 = var.tags
}

module "aca_env" {
  source                      = "../modules/containerapps_env"
  name                        = var.config.env_name
  location                    = var.location
  resource_group_name         = module.rg.name
  subnet_id                   = module.net.subnet_id
  logs_destination            = "azure-monitor"
  log_analytics_workspace_id  = module.obs.log_analytics_id
  tags                        = var.tags
}

module "env_diag" {
  source                      = "../modules/diagnostic_setting"
  name                        = "${var.config.env_name}-diag"
  target_resource_id          = module.aca_env.id
  log_analytics_workspace_id  = module.obs.log_analytics_id
  category_groups             = ["allLogs"]
  enable_metrics              = true
}

module "cosmos" {
  source                    = "../modules/cosmos_mongo"
  name                      = var.config.cosmos_name
  location                  = var.location
  resource_group_name       = module.rg.name
  create_database           = true
  database_name             = "appdb"
  database_throughput       = var.config.database_throughput
  subnet_id                 = module.net.subnet_id
  allowed_ip_ranges         = []
  log_analytics_workspace_id = module.obs.log_analytics_id
  tags                      = var.tags
}

module "apps" {
  for_each = var.config.apps
  source   = "../modules/containerapp"

  name                        = each.key
  resource_group_name         = module.rg.name
  environment_id              = module.aca_env.id
  image                       = each.value.image
  target_port                 = each.value.port
  cpu                         = each.value.cpu
  memory                      = each.value.memory
  min_replicas                = each.value.min_replicas
  max_replicas                = each.value.max_replicas
  enable_system_identity      = true
  log_analytics_workspace_id  = module.obs.log_analytics_id
  environment_variables       = local.environment_variables
  http_scale_rule             = each.value.http_scale_rule
  cpu_scale_rule              = each.value.cpu_scale_rule
  memory_scale_rule           = each.value.memory_scale_rule
  azure_queue_scale_rules     = each.value.azure_queue_scale_rules
  custom_scale_rules          = each.value.custom_scale_rules
  tags                        = merge(var.tags, { app = each.key })
}

module "alerts" {
  count = var.config.enable_alerts ? 1 : 0
  
  source               = "../modules/alerts"
  name                 = "${var.environment}-alerts"
  short_name           = var.environment
  resource_group_name  = module.rg.name
  email_receivers      = var.alert_emails
  container_app_ids    = { for k, m in module.apps : k => m.id }
  requests_5xx_threshold = var.config.requests_5xx_threshold
  tags                 = var.tags
}

data "azurerm_client_config" "current" {}

# Cost-optimized security - essential security at minimal cost
module "cost_optimized_security" {
  count = var.config.enable_security_features ? 1 : 0
  
  source                     = "../modules/cost_optimized_security"
  name_prefix                = var.environment
  location                   = var.location
  resource_group_name        = module.rg.name
  subnet_id                  = module.net.subnet_id
  enable_key_vault           = true
  key_vault_name             = var.config.key_vault_name != null ? var.config.key_vault_name : "${var.config.cosmos_name}-kv"
  key_vault_admins           = [data.azurerm_client_config.current.object_id]
  container_app_identities   = { for k, m in module.apps : k => m.identity_principal_id }
  cosmos_connection_string   = module.cosmos.connection_string
  enable_basic_monitoring    = var.config.enable_alerts
  container_app_ids          = { for k, m in module.apps : k => m.id }
  action_group_id            = var.config.enable_alerts ? module.alerts[0].action_group_id : null
  log_analytics_workspace_id = module.obs.log_analytics_id
  tags                       = var.tags
}