variable "name" { type = string }
variable "short_name" { type = string }
variable "resource_group_name" { type = string }
variable "email_receivers" {
  type    = map(string)
  default = {}
}
variable "tags" {
  type    = map(string)
  default = {}
}