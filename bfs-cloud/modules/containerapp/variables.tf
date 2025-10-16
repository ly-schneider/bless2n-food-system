variable "name" { type = string }
variable "resource_group_name" { type = string }
variable "environment_id" { type = string }
variable "image" { type = string }
variable "target_port" { type = number }
variable "external_ingress" {
  type        = bool
  description = "Whether ingress is exposed externally (true) or internal-only (false)"
  default     = true
}
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

variable "secrets" {
  type        = map(string)
  description = "Container App secrets (name => value), e.g., GHCR tokens"
  default     = {}
}

variable "registries" {
  type = list(object({
    server               = string
    username             = string
    password_secret_name = string
  }))
  description = "Container registries for pulling private images"
  default     = []
}
variable "enable_system_identity" {
  type    = bool
  default = true
}

variable "user_assigned_identity_ids" {
  type        = list(string)
  description = "List of user-assigned managed identity IDs"
  default     = []
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

variable "key_vault_secrets" {
  type        = map(string)
  description = "Key Vault secrets to inject as environment variables (env_var_name => secret_name)"
  default     = {}
}

variable "key_vault_secret_refs" {
  type        = map(string)
  description = "Key Vault secret references with full resource IDs (secret_name => key_vault_secret_id)"
  default     = {}
}

variable "health_check_path" {
  type        = string
  description = "Health check path for liveness and readiness probes"
  default     = "/health"
}

variable "read_only_filesystem" {
  type        = bool
  description = "Use read-only root filesystem for security"
  default     = true
}

variable "required_capabilities" {
  type        = list(string)
  description = "Required Linux capabilities"
  default     = []
}

variable "volume_mounts" {
  type = list(object({
    name = string
    path = string
  }))
  description = "Volume mounts for the container"
  default     = []
}

variable "volumes" {
  type = list(object({
    name         = string
    storage_type = string
    storage_name = string
  }))
  description = "Volumes for the container app"
  default     = []
}
