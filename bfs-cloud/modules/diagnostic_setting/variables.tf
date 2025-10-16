variable "name" { type = string }
variable "target_resource_id" { type = string }
variable "log_analytics_workspace_id" { type = string }
variable "categories" {
  type    = list(string)
  default = []
}
variable "category_groups" {
  type    = list(string)
  default = []
}
variable "enable_metrics" {
  type    = bool
  default = true
}