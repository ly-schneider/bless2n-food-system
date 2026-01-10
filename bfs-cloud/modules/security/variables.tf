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

variable "enable_key_vault" {
  description = "Enable Key Vault data source lookup"
  type        = bool
  default     = true
}

variable "key_vault_name" {
  description = "Key Vault name"
  type        = string
}

variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}
