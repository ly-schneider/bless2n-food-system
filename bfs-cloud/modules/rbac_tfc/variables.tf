variable "target_rg_id" {
  description = "Resource Group scope for baseline or RG-scoped assignments"
  type        = string
}

variable "principal_client_id" {
  description = "Optional: Client ID (appId) of the service principal to assign roles to. If null, uses the current caller."
  type        = string
  default     = null
}

variable "principal_object_id" {
  description = "Optional: Object ID of the principal to assign roles to. Overrides client_id/current."
  type        = string
  default     = null
}

variable "network_scopes" {
  description = "Scopes (subnet/VNet/resource group) to grant Network Contributor"
  type        = list(string)
  default     = []
}

variable "private_dns_zone_scopes" {
  description = "Scopes (Private DNS zone IDs or RG IDs) to grant Private DNS Zone Contributor"
  type        = list(string)
  default     = []
}

variable "uaa_scopes" {
  description = "Scopes to grant User Access Administrator (for Terraform to manage role assignments)"
  type        = list(string)
  default     = []
}

variable "managed_identity_scopes" {
  description = "Scopes to grant Managed Identity Contributor (if Terraform creates user-assigned identities)"
  type        = list(string)
  default     = []
}

variable "cosmos_account_scopes" {
  description = "Cosmos DB account scopes to grant management or key-read roles"
  type        = list(string)
  default     = []
}

variable "grant_cosmos_account_contributor" {
  description = "Grant Cosmos DB Account Contributor on cosmos_account_scopes (for create/update)"
  type        = bool
  default     = true
}

variable "grant_cosmos_keys_reader" {
  description = "Grant minimal reader for cosmos keys on cosmos_account_scopes (uses built-in 'Cosmos DB Account Reader Role')"
  type        = bool
  default     = true
}

variable "baseline_enable_contributor_on_rg" {
  description = "Fallback: also grant Contributor at RG scope"
  type        = bool
  default     = false
}

variable "baseline_enable_uaa_on_rg" {
  description = "Fallback: also grant User Access Administrator at RG scope"
  type        = bool
  default     = false
}

