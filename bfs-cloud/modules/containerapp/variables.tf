variable "name"                        { type = string }
variable "resource_group_name"         { type = string }
variable "environment_id"              { type = string }
variable "image"                       { type = string }
variable "target_port"                 { type = number }
variable "cpu" {
  type    = number
  default = 0.5
}
variable "memory" {
  type    = string
  default = "1.0Gi"
}
variable "min_replicas" {
  type    = number
  default = 1
}
variable "max_replicas" {
  type    = number
  default = 2
}
variable "environment_variables" {
  type    = map(string)
  default = {}
}
variable "enable_system_identity" {
  type    = bool
  default = true
}
variable "log_analytics_workspace_id" {
  type    = string
  default = null
}
variable "tags" {
  type    = map(string)
  default = {}
}