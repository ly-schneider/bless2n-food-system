variable "name"                 { type = string }
variable "location"             { type = string }
variable "resource_group_name"  { type = string }
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
  description = "Subnet ID for VNet integration"
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
variable "disable_local_auth" {
  description = "Disable local authentication and use AAD only"
  type        = bool
  default     = false
}
variable "log_analytics_workspace_id" {
  description = "Log Analytics workspace ID for diagnostics"
  type        = string
}
variable "tags" {
  type    = map(string)
  default = {}
}