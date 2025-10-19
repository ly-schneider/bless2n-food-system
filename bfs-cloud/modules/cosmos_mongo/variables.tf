variable "name" { type = string }
variable "location" { type = string }
variable "resource_group_name" { type = string }
variable "create_database" {
  type    = bool
  default = false
}
variable "database_name" {
  type    = string
  default = "appdb"
}
variable "database_throughput" {
  type    = number
  default = 400
}
variable "subnet_id" {
  description = "Subnet ID for VNet integration (not used for Private Endpoint)"
  type        = string
}
variable "enable_private_endpoint" {
  description = "Whether to create a Private Endpoint and Private DNS for Cosmos"
  type        = bool
  default     = false
}
variable "private_endpoint_subnet_id" {
  description = "Subnet ID to place the Cosmos DB Private Endpoint (must NOT be delegated)"
  type        = string
}
variable "vnet_id" {
  description = "Virtual Network ID for Private DNS zone link"
  type        = string
}
variable "allowed_ip_ranges" {
  description = "Allowed IP ranges for Cosmos DB access"
  type        = list(string)
  default     = []
}
variable "cors_allowed_origins" {
  description = "CORS allowed origins"
  type        = list(string)
  default     = ["*"]
}
variable "log_analytics_workspace_id" {
  description = "Log Analytics workspace ID for diagnostics"
  type        = string
}
variable "tags" {
  type    = map(string)
  default = {}
}
