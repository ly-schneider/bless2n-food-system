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

variable "revision_suffix" {
  description = "Unique suffix to force a new Container Apps revision (e.g., commit SHA)"
  type        = string
  default     = null
}

variable "app_secrets" {
  description = "Map of app name => map of secret name => value"
  type        = map(map(string))
  default     = {}
}
