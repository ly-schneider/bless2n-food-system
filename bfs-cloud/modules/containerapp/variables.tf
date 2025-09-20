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

variable "revision_suffix" {
  type        = string
  description = "Revision suffix for the container app"
  default     = null
}

variable "http_scale_rule" {
  type = object({
    name                = string
    concurrent_requests = number
  })
  description = "HTTP scale rule configuration"
  default     = null
}

variable "cpu_scale_rule" {
  type = object({
    name           = string
    cpu_percentage = number
  })
  description = "CPU scale rule configuration"
  default     = null
}

variable "memory_scale_rule" {
  type = object({
    name              = string
    memory_percentage = number
  })
  description = "Memory scale rule configuration"
  default     = null
}

variable "azure_queue_scale_rules" {
  type = list(object({
    name              = string
    queue_name        = string
    queue_length      = number
    secret_name       = string
    trigger_parameter = string
  }))
  description = "Azure Queue scale rules configuration"
  default     = []
}

variable "custom_scale_rules" {
  type = list(object({
    name             = string
    custom_rule_type = string
    metadata         = map(string)
    authentication = optional(object({
      secret_name       = string
      trigger_parameter = string
    }))
  }))
  description = "Custom scale rules configuration"
  default     = []
}