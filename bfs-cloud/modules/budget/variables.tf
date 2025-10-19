variable "name" { type = string }
variable "resource_group_id" { type = string }
variable "amount" { type = number }
variable "time_grain" {
  type    = string
  default = "Monthly"
}
variable "start_date" {
  type    = string
  default = "2025-01-01T00:00:00Z"
}
variable "action_group_id" {
  type    = string
  default = null
}
variable "tags" {
  type    = map(string)
  default = {}
}
