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

# Optional digests published by the image build (prefer over tag when set)
variable "frontend_digest" {
  description = "OCI digest for the frontend image (e.g., sha256:...)"
  type        = string
  default     = ""
}

variable "backend_digest" {
  description = "OCI digest for the backend image (e.g., sha256:...)"
  type        = string
  default     = ""
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
