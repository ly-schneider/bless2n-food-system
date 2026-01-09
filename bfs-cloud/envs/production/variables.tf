# Deployment-specific variables that can't be centralized

variable "image_tag" {
  description = "Tag to use for images (commit SHA, tag, or branch)"
  type        = string
  default     = "production"
}

variable "frontend_digest" {
  description = "OCI digest for the frontend image (e.g., sha256:...). Preferred over tag when set."
  type        = string
  default     = ""
}

variable "backend_digest" {
  description = "OCI digest for the backend image (e.g., sha256:...). Preferred over tag when set."
  type        = string
  default     = ""
}

variable "revision_suffix" {
  description = "Unique suffix to force a new Container Apps revision (e.g., commit SHA)"
  type        = string
  default     = null
}

variable "ghcr_token" {
  description = "GitHub Container Registry personal access token for pulling images"
  type        = string
  sensitive   = true
}
