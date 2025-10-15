variable "location" {
  type    = string
  default = "northeurope"
}
variable "tags" {
  type    = map(string)
  default = { project = "bfs", env = "staging" }
}

variable "alert_emails" {
  type    = map(string)
  default = {}
}

# CI/CD-friendly container registry and image settings
variable "registry_server" {
  description = "Container registry server hostname (e.g., ghcr.io)"
  type        = string
  default     = "ghcr.io"
}

variable "registry_namespace" {
  description = "Registry namespace/owner (e.g., ghcr.io/<owner>)"
  type        = string
  default     = "example"
}

variable "registry_username" {
  description = "Username for private registry auth (e.g., GitHub username)"
  type        = string
  default     = null
}

variable "registry_token" {
  description = "Token/password for private registry auth (e.g., GHCR PAT)"
  type        = string
  sensitive   = true
  default     = null
}

# ACR options for staging/prod simplicity
variable "enable_acr" {
  description = "Use Azure Container Registry for images and managed identity pull"
  type        = bool
  default     = true
}

variable "acr_name" {
  description = "Azure Container Registry name (without .azurecr.io)"
  type        = string
  default     = "bfsstagingacr"
}

variable "image_tag" {
  description = "Tag to use for images (commit SHA, tag, or branch)"
  type        = string
  default     = "staging"
}

# Optional per-app overrides for secrets and registries
variable "app_secrets" {
  description = "Map of app name => map of secret name => value"
  type        = map(map(string))
  default     = {}
}

variable "app_registries" {
  description = "Map of app name => list of registry credentials"
  type = map(list(object({
    server               = string
    username             = string
    password_secret_name = string
  })))
  default = {}
}
