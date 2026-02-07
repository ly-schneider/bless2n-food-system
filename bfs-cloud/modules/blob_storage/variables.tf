variable "name" {
  description = "Storage account name (must be globally unique, 3-24 lowercase alphanumeric)"
  type        = string
}

variable "resource_group_name" {
  description = "Resource group to create resources in"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}

variable "container_name" {
  description = "Blob container name"
  type        = string
  default     = "product-images"
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}
