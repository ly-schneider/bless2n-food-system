variable "name" {
  description = "Name of the Azure Container Registry"
  type        = string
}

variable "resource_group_name" {
  description = "Name of the resource group where ACR will be created"
  type        = string
}

variable "location" {
  description = "Azure region where ACR will be created"
  type        = string
}

variable "sku" {
  description = "SKU for the Azure Container Registry (Basic, Standard, Premium)"
  type        = string
  default     = "Basic"
}

variable "admin_enabled" {
  description = "Whether admin user is enabled for the ACR"
  type        = bool
  default     = false
}