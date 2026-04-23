variable "grafana_stack_url" {
  description = "Grafana Cloud stack URL (e.g. https://leysservices.grafana.net)"
  type        = string
  default     = "https://leysservices.grafana.net"
}

variable "grafana_cloud_token" {
  description = "Grafana Cloud access policy token (scopes: dashboards:write, folders:write, datasources:write, alerts:write)"
  type        = string
  sensitive   = true
}

variable "azure_tenant_id" {
  description = "Azure AD tenant ID for the Grafana Azure Monitor data source"
  type        = string
}

variable "azure_subscription_id" {
  description = "Azure subscription ID containing the bfs staging + production resource groups"
  type        = string
}

variable "grafana_azure_client_id" {
  description = "Client ID (appId) of the pre-created sp-grafana-cloud-reader service principal"
  type        = string
}

variable "grafana_azure_client_secret" {
  description = "Client secret of sp-grafana-cloud-reader"
  type        = string
  sensitive   = true
}

variable "sentry_organization_slug" {
  description = "Sentry organization slug"
  type        = string
  default     = "leys"
}

variable "sentry_auth_token" {
  description = "Sentry internal integration token with org:read + project:read + event:read scopes. Leave empty to skip Sentry wiring."
  type        = string
  sensitive   = true
  default     = ""
}

variable "staging_remote_state_workspace" {
  description = "TF Cloud workspace name that owns staging Azure infra"
  type        = string
  default     = "bfs-staging"
}

variable "production_remote_state_workspace" {
  description = "TF Cloud workspace name that owns production Azure infra"
  type        = string
  default     = "bfs-production"
}

variable "neon_pg_host_staging" {
  description = "Neon Postgres host for the staging branch (e.g. ep-xxxx-staging.eu-central-1.aws.neon.tech). Leave empty to skip the staging Postgres data source."
  type        = string
  default     = ""
}

variable "neon_pg_host_production" {
  description = "Neon Postgres host for the production branch. Leave empty to skip the production Postgres data source."
  type        = string
  default     = ""
}

variable "neon_pg_database" {
  description = "Neon Postgres database name that holds BFS tables"
  type        = string
  default     = "bfs"
}

variable "neon_grafana_password" {
  description = "Password for the grafana_reader role on Neon. Shared between staging and production branches."
  type        = string
  sensitive   = true
  default     = ""
}
