variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}

variable "resource_group_name" {
  description = "Resource group name"
  type        = string
}

variable "subnet_id" {
  description = "Container Apps subnet ID"
  type        = string
}

variable "enable_key_vault" {
  description = "Enable basic Key Vault (Standard SKU)"
  type        = bool
  default     = true
}

variable "key_vault_name" {
  description = "Key Vault name"
  type        = string
}

variable "allowed_ip_ranges" {
  description = "Allowed IP ranges for Key Vault access"
  type        = list(string)
  default     = []
}

variable "key_vault_admins" {
  description = "Object IDs of Key Vault administrators"
  type        = list(string)
  default     = []
}

variable "uami_principal_id" {
  description = "User-assigned managed identity principal ID for container apps"
  type        = string
}

variable "cosmos_connection_string" {
  description = "Cosmos DB connection string"
  type        = string
  sensitive   = true
  default     = ""
}

variable "enable_basic_monitoring" {
  description = "Enable basic monitoring alerts"
  type        = bool
  default     = true
}

variable "container_app_ids" {
  description = "Map of container app names to their resource IDs"
  type        = map(string)
  default     = {}
}

variable "action_group_id" {
  description = "Action group ID for alerts"
  type        = string
  default     = null
}

variable "log_analytics_workspace_id" {
  description = "Log Analytics workspace ID for basic logging"
  type        = string
}

variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}
