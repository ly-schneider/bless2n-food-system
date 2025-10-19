variable "name" {
  type = string
}

variable "short_name" {
  type = string
}

variable "resource_group_name" {
  type = string
}

variable "email_receivers" {
  type    = map(string)
  default = {}
}

variable "action_group_id" {
  description = "Use an existing Action Group instead of creating a new one"
  type        = string
  default     = null
}

variable "container_app_ids" {
  type = map(string)
}

variable "requests_5xx_threshold" {
  type    = number
  default = 10
}

variable "tags" {
  type    = map(string)
  default = {}
}
