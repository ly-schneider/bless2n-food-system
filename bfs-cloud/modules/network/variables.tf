variable "resource_group_name" { type = string }
variable "location" { type = string }
variable "vnet_name" { type = string }
variable "vnet_cidr" { type = string }
variable "subnet_name" { type = string }
variable "subnet_cidr" { type = string }
variable "workload_profiles_subnet_name" {
  description = "Subnet name for Container Apps Workload Profiles (must NOT be delegated)"
  type        = string
  default     = "aca-workload-subnet"
}
variable "workload_profiles_subnet_cidr" {
  description = "Address prefix for Workload Profiles subnet (non-delegated)"
  type        = string
  default     = "10.1.9.0/24"
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
