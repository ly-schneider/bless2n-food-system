variable "name" {
  description = "Name of the Cosmos DB account"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}

variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
}

variable "create_database" {
  description = "Whether to create a MongoDB database"
  type        = bool
  default     = false
}

variable "database_name" {
  description = "Name of the MongoDB database"
  type        = string
  default     = "appdb"
}

variable "database_throughput" {
  description = "Throughput for the database (RU/s)"
  type        = number
  default     = 400
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
  description = "Resource tags"
  type        = map(string)
  default     = {}
}
