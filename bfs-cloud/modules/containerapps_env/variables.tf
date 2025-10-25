variable "name" { type = string }
variable "location" { type = string }
variable "resource_group_name" { type = string }
variable "subnet_id" { type = string }
variable "workload_profile_subnet_id" {
  description = "Non-delegated subnet ID for Container Apps Workload Profiles (agent pool). Leave null to omit."
  type        = string
  default     = null
}
variable "logs_destination" {
  type    = string
  default = "azure-monitor"
}
variable "log_analytics_workspace_id" {
  type    = string
  default = null
}
variable "tags" {
  type    = map(string)
  default = {}
}
