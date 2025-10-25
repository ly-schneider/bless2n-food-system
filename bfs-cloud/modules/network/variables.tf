variable "resource_group_name" { type = string }
variable "location" { type = string }
variable "vnet_name" { type = string }
variable "vnet_cidr" { type = string }
variable "subnet_name" { type = string }
variable "subnet_cidr" { type = string }
variable "delegate_containerapps_subnet" {
  description = "If true, delegate the ACA subnet to Microsoft.App/environments (Consumption). For Workload Profiles, set false."
  type        = bool
  default     = false
}
variable "private_endpoints_subnet_name" {
  description = "Subnet name for Private Endpoints"
  type        = string
  default     = "private-endpoints-subnet"
}
variable "private_endpoints_subnet_cidr" {
  description = "Address prefix for Private Endpoints subnet"
  type        = string
  default     = "10.1.8.0/24"
}
variable "tags" {
  type    = map(string)
  default = {}
}
