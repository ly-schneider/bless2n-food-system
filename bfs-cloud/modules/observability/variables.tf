variable "resource_group_name" { type = string }
variable "location" { type = string }
variable "law_name" { type = string }
variable "appi_name" { type = string }
variable "retention_days" {
  type    = number
  default = 30
}
variable "enable_app_insights" {
  type    = bool
  default = false
}
variable "tags" {
  type    = map(string)
  default = {}
}